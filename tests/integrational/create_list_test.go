package integrational

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func InitTestDB() *sql.DB {
	connStr := "user=postgres dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func InitTestRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return client
}

type Repository struct {
	DB       *sql.DB
	Redis    *redis.Client
	TodoList *TodoListRepository
	TodoItem *TodoItemRepository
}

func NewRepository(db *sql.DB, redisClient *redis.Client) *Repository {
	return &Repository{
		DB:       db,
		Redis:    redisClient,
		TodoList: NewTodoListRepository(db),
		TodoItem: NewTodoItemRepository(db),
	}
}

type TodoListRepository struct {
	DB *sql.DB
}

func NewTodoListRepository(db *sql.DB) *TodoListRepository {
	return &TodoListRepository{
		DB: db,
	}
}

type TodoList struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (r *TodoListRepository) Create(userID int, list todo.TodoList) (int, error) {
	query := `INSERT INTO todo_lists (user_id, title) VALUES (?, ?)`
	result, err := r.DB.Exec(query, userID, list.Title)
	if err != nil {
		log.Println("Error inserting todo list:", err)
		return 0, err
	}

	listID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error retrieving last insert ID:", err)
		return 0, err
	}

	return int(listID), nil
}

type Service struct {
	Authorization *service.AuthService
	TodoList      *service.TodoListService
	TodoItem      *service.TodoItemService
}

func NewService(repo *Repository) *Service {
	return &Service{
		Authorization: &service.AuthService{},
		TodoList:      &service.TodoListService{},
		TodoItem:      &service.TodoItemService{},
	}
}

type Handler struct {
	Authorization *service.AuthService
	TodoList      *service.TodoListService
	TodoItem      *service.TodoItemService
}

func InitRoutes(router *gin.Engine, handler *Handler) {
	router.POST("/auth/sign-up", handler.SignUpHandler)
	router.POST("/auth/sign-in", handler.SignInHandler)
	router.POST("/api/lists", handler.CreateListHandler)
}

func (h *Handler) SignUpHandler(c *gin.Context) {
	var user todo.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, err := h.Authorization.CreateUser(todo.User{Username: user.Username, Password: user.Password})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID})
}

func (h *Handler) SignInHandler(c *gin.Context) {
	var user todo.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	token, err := h.Authorization.GenerateToken(user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) CreateListHandler(c *gin.Context) {
	var list todo.TodoList
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := 1
	listID, err := h.TodoList.CreateList(userID, list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": listID})
}

func getAuthTokenForIntegrationTest(router *gin.Engine) string {
	body := `{"username":"testuser","password":"password"}`
	req, _ := http.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["token"]
}

func NewHandler(authorization *service.AuthService, todoList *service.TodoListService, todoItem *service.TodoItemService) *Handler {
	return &Handler{
		Authorization: authorization,
		TodoList:      todoList,
		TodoItem:      todoItem,
	}
}

func TestCreateListIntegration(t *testing.T) {
	db := InitTestDB()
	redisClient := InitTestRedis()
	repo := NewRepository(db, redisClient)
	service := NewService(repo)
	handler := NewHandler(service.Authorization, service.TodoList, service.TodoItem)

	router := gin.Default()
	InitRoutes(router, handler)

	token := getAuthTokenForIntegrationTest(router)

	body := `{"title":"Integration Test List"}`
	req, _ := http.NewRequest("POST", "/api/lists", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["id"])
}
