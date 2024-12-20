package module

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getAuthToken() string {
	body := `{"username":"testuser","password":"password"}`
	req, _ := http.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := setupTestRouter()
	router.ServeHTTP(w, req)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["token"]
}

func TestCreateListFunctionality(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken()

	body := `{"title":"My Task List"}`
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
