package functional

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func createUser(username, password string) (int, error) {
	return 1, nil
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := createUser(user.Username, user.Password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{"id": userID}
	json.NewEncoder(w).Encode(response)
}

func setupTestRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/auth/sign-up", SignUpHandler).Methods("POST")
	return router
}

func TestSignInFunctionality(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"testuser","password":"password"}`
	req, _ := http.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}
