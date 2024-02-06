package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/aryan9600/service-catalog/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		name       string
		body       UserAuthInput
		assertFunc func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "register a new user",
			body: UserAuthInput{
				Username: "bob",
				Password: "secret",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 201, w.Code)

				var response RegisterOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				assert.Equal(t, response.Data.Username, "bob")
			},
		},
		{
			name: "register an existing user",
			body: UserAuthInput{
				Username: "user1",
				Password: "secret",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 400, w.Code)
				assert.Contains(t, string(w.Body.Bytes()), models.ErrUniqueConstraintViolation.Error())
			},
		},
		{
			name: "register an user with a username > 20 chars",
			body: UserAuthInput{
				Username: getStr(21),
				Password: "secret",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 400, w.Code)
				assert.Contains(t, string(w.Body.Bytes()), "invalid registration input")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			tt.assertFunc(t, w)
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       UserAuthInput
		assertFunc func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "logging in via valid creds",
			body: UserAuthInput{
				Username: "user1",
				Password: "pwd1",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)
				var response LoginOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				token := response.AccessToken
				err = auth.CheckTokenValidity(token)
				assert.NoError(t, err)
			},
		},
		{
			name: "logging in via invalid username",
			body: UserAuthInput{
				Username: "whodis",
				Password: "pwd1",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 404, w.Code)
				assert.Contains(t, string(w.Body.Bytes()), models.ErrRecordNotFound.Error())
			},
		},
		{
			name: "logging in via invalid password",
			body: UserAuthInput{
				Username: "user1",
				Password: "wrongpwd",
			},
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 401, w.Code)
				assert.Contains(t, string(w.Body.Bytes()), "invalid password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			tt.assertFunc(t, w)
		})
	}
}

func getStr(n int) string {
	str := ""
	for i := 0; i < n; i++ {
		str += fmt.Sprintf("%da", i)
	}
	return str
}
