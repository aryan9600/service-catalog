package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aryan9600/service-catalog/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestListServices(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		auth       bool
		userID     uint
		assertFunc func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "listing services for an authenticated user returns a list of services",
			path:   "/services",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)

				var response ListServicesOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				assert.Len(t, response.Data, 3)
				for _, svc := range response.Data {
					if svc.UserID != 1 {
						t.Fatalf("Unexpected data; requested services for user id %d; got services for user id %d", 1, svc.UserID)
					}
				}
			},
		},
		{
			name:   "listing services with a limit and offset",
			path:   "/services?limit=2&offset=1",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)

				var response ListServicesOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				assert.Len(t, response.Data, 2)
				for _, svc := range response.Data {
					if svc.UserID != 1 {
						t.Fatalf("Unexpected data; requested services for user id %d; got services for user id %d", 1, svc.UserID)
					}
				}
				// The first object should be the second record since offset=1
				assert.Equal(t, uint(2), response.Data[0].ID)
			},
		},
		{
			name:   "listing services sorted by name in a descending order",
			path:   "/services?sortKey=name&descending=true",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)

				var response ListServicesOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Len(t, response.Data, 3)

				var names []string
				for _, svc := range response.Data {
					if svc.UserID != 1 {
						t.Fatalf("Unexpected data; requested services for user id %d; got services for user id %d", 1, svc.UserID)
					}
					names = append(names, svc.Name)
				}
				assert.Equal(t, []string{"storage", "dns", "auth"}, names)
			},
		},
		{
			name:   "listing services filtered by name",
			path:   "/services?name=mesh",
			auth:   true,
			userID: uint(2),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)

				var response ListServicesOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Len(t, response.Data, 1)
				service := response.Data[0]
				assert.Equal(t, 2, service.UserID)
				assert.Equal(t, "service mesh", service.Name)
			},
		},
		{
			name: "listing services for an unauthenticated user returns a 401",
			path: "/services",
			auth: false,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 401, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)
			if tt.auth {
				err = addAuthorizationHeader(tt.userID, req)
				assert.NoError(t, err)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			tt.assertFunc(t, w)
		})
	}
}

func TestGetService(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		auth       bool
		userID     uint
		assertFunc func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "fetch a service for an authenticated user",
			path:   "/services/1",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)
				var response ServiceOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Equal(t, uint(1), response.Data.ID)
			},
		},
		{
			name: "fetch a service for an unauthenticated user",
			path: "/services/1",
			auth: false,
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 401, w.Code)
			},
		},
		{
			name:   "fetch a unrelated service for an authenticated user",
			path:   "/services/4",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 404, w.Code)
			},
		},
		{
			name:   "fetch a service with versions for an authenticated user",
			path:   "/services/1?versions=true",
			auth:   true,
			userID: uint(1),
			assertFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 200, w.Code)
				var response GetServiceWithVersionsOutput
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Equal(t, uint(1), response.Data.ID)
				assert.Len(t, response.Data.Versions, 2)
				assert.Equal(t, response.Data.Versions[0].Version, "1.0")
				assert.Equal(t, response.Data.Versions[1].Version, "1.1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)
			if tt.auth {
				err = addAuthorizationHeader(tt.userID, req)
				assert.NoError(t, err)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			tt.assertFunc(t, w)
		})
	}
}

func addAuthorizationHeader(userID uint, req *http.Request) error {
	token, err := auth.GenerateToken(userID)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return nil
}
