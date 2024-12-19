package service

import (
	"crypto/sha1"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	salt       = "47nv74bv6hknjk443h8ewf"
	signingKey = "js8u5hds12g058%7G5&"
)

type TokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user todo.User) (int, error) {
	if user.Password == "" {
		return 0, fmt.Errorf("password cannot be empty")
	}

	user.Password = GeneratePasswordHash(user.Password)
	return s.repo.CreateUser(user)
}

func GeneratePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func (s *AuthService) GenerateToken(username, password string) (string, error) {
	user, err := s.repo.GetUser(username, GeneratePasswordHash(password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
		UserId: user.Id,
	})

	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) ParseToken(token string) (int, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signing method: %v", token.Header["alg"])
		}

		return []byte(signingKey), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := parsedToken.Claims.(*TokenClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	return claims.UserId, nil
}

func (s *AuthService) GetUser(username, password string) (todo.User, error) {
	if username == "" {
		return todo.User{}, fmt.Errorf("username cannot be empty")
	} else if password == "" {
		return todo.User{}, fmt.Errorf("password cannot be empty")
	}

	password = GeneratePasswordHash(password)
	return s.repo.GetUser(username, password)
}

func (s *AuthService) UpdateUser(userId int, user todo.User) error {
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	} else if user.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	user.Password = GeneratePasswordHash(user.Password)
	return s.repo.UpdateUser(userId, user)
}

func (s *AuthService) DeleteUser(userId int) error {
	return s.repo.DeleteUser(userId)
}
