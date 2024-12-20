package module

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllListsFunctionality(t *testing.T) {
	router := setupTestRouter()
	token := getAuthToken()

	req, _ := http.NewRequest("GET", "/api/lists", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.GreaterOrEqual(t, len(response), 1)
}
