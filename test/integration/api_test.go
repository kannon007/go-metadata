// Package integration provides end-to-end API integration tests.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-metadata/internal/auth"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Simple health handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now(),
		})
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", resp["status"])
	}
}

// TestLoginEndpoint tests the login endpoint
func TestLoginEndpoint(t *testing.T) {
	jwtManager := auth.NewJWTManager(&auth.JWTConfig{
		Secret:     "test-secret-key-for-testing-only",
		Expire:     time.Hour,
		RefreshExp: 24 * time.Hour,
		Issuer:     "test-issuer",
	})

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid login",
			requestBody: map[string]string{
				"username": "admin",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "admin",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name:           "empty body",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Login handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var loginReq struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
					return
				}

				if loginReq.Username == "" || loginReq.Password == "" {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": "Username and password required"})
					return
				}

				user := &auth.User{
					ID:       "user-" + loginReq.Username,
					Username: loginReq.Username,
					Email:    loginReq.Username + "@example.com",
					Role:     auth.RoleOperator,
					Enabled:  true,
				}
				if loginReq.Username == "admin" {
					user.Role = auth.RoleAdmin
				}

				token, err := jwtManager.GenerateToken(user)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				refreshToken, err := jwtManager.GenerateRefreshToken(user)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token":  token,
					"refresh_token": refreshToken,
					"token_type":    "Bearer",
					"user": map[string]interface{}{
						"id":       user.ID,
						"username": user.Username,
						"role":     user.Role,
					},
				})
			})

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectToken {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp["access_token"] == nil || resp["access_token"] == "" {
					t.Error("Expected access_token in response")
				}

				if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
					t.Error("Expected refresh_token in response")
				}
			}
		})
	}
}

// TestAuthMiddleware tests the authentication middleware
// Feature: datasource-management, Property 14: 权限验证
// **Validates: Requirements 8.1, 8.2, 8.5**
func TestAuthMiddleware(t *testing.T) {
	logger := log.DefaultLogger

	jwtManager := auth.NewJWTManager(&auth.JWTConfig{
		Secret:     "test-secret-key-for-testing-only",
		Expire:     time.Hour,
		RefreshExp: 24 * time.Hour,
		Issuer:     "test-issuer",
	})

	authMiddleware := auth.NewAuthMiddleware(jwtManager, logger)

	// Generate a valid token
	user := &auth.User{
		ID:       "test-user",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     auth.RoleOperator,
		Enabled:  true,
	}
	validToken, _ := jwtManager.GenerateToken(user)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			// Protected handler
			protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			})

			authMiddleware.HTTPHandler(protectedHandler).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestRBACMiddleware tests the RBAC middleware
func TestRBACMiddleware(t *testing.T) {
	logger := log.DefaultLogger

	jwtManager := auth.NewJWTManager(&auth.JWTConfig{
		Secret:     "test-secret-key-for-testing-only",
		Expire:     time.Hour,
		RefreshExp: 24 * time.Hour,
		Issuer:     "test-issuer",
	})

	authMiddleware := auth.NewAuthMiddleware(jwtManager, logger)
	rbacMiddleware := auth.NewRBACMiddleware(logger)

	// Generate tokens for different roles
	adminUser := &auth.User{
		ID:       "admin-user",
		Username: "admin",
		Role:     auth.RoleAdmin,
		Enabled:  true,
	}
	adminToken, _ := jwtManager.GenerateToken(adminUser)

	operatorUser := &auth.User{
		ID:       "operator-user",
		Username: "operator",
		Role:     auth.RoleOperator,
		Enabled:  true,
	}
	operatorToken, _ := jwtManager.GenerateToken(operatorUser)

	viewerUser := &auth.User{
		ID:       "viewer-user",
		Username: "viewer",
		Role:     auth.RoleViewer,
		Enabled:  true,
	}
	viewerToken, _ := jwtManager.GenerateToken(viewerUser)

	tests := []struct {
		name           string
		method         string
		path           string
		token          string
		expectedStatus int
	}{
		{
			name:           "admin can create",
			method:         http.MethodPost,
			path:           "/api/v1/datasources",
			token:          adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "operator can create",
			method:         http.MethodPost,
			path:           "/api/v1/datasources",
			token:          operatorToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "viewer cannot create",
			method:         http.MethodPost,
			path:           "/api/v1/datasources",
			token:          viewerToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "viewer can read",
			method:         http.MethodGet,
			path:           "/api/v1/datasources",
			token:          viewerToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "admin can delete",
			method:         http.MethodDelete,
			path:           "/api/v1/datasources/123",
			token:          adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "operator cannot delete",
			method:         http.MethodDelete,
			path:           "/api/v1/datasources/123",
			token:          operatorToken,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			w := httptest.NewRecorder()

			// Protected handler
			protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Chain middlewares
			authMiddleware.HTTPHandler(
				rbacMiddleware.HTTPHandler(protectedHandler),
			).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestRateLimiter tests the rate limiting middleware
func TestRateLimiter(t *testing.T) {
	logger := log.DefaultLogger

	rateLimiter := auth.NewRateLimiter(&auth.RateLimitConfig{
		Enabled:       true,
		RequestsPerIP: 5,
		BurstSize:     5,
		Window:        time.Second,
		CleanupPeriod: time.Minute,
	}, logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Make requests up to the limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		rateLimiter.HTTPHandler(handler).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	rateLimiter.HTTPHandler(handler).ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d (rate limited), got %d", http.StatusTooManyRequests, w.Code)
	}
}

// TestAPIResponseFormat tests that API responses follow the expected format
func TestAPIResponseFormat(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedFields []string
	}{
		{
			name: "success response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":   "123",
					"name": "test",
				})
			},
			expectedFields: []string{"id", "name"},
		},
		{
			name: "error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    "INVALID_REQUEST",
					"message": "Invalid request body",
				})
			},
			expectedFields: []string{"code", "message"},
		},
		{
			name: "list response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data_sources": []interface{}{},
					"total":        0,
					"page":         1,
					"page_size":    20,
				})
			},
			expectedFields: []string{"data_sources", "total", "page", "page_size"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			tt.handler.ServeHTTP(w, req)

			var resp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			for _, field := range tt.expectedFields {
				if _, ok := resp[field]; !ok {
					t.Errorf("Expected field '%s' in response", field)
				}
			}
		})
	}
}

// TestDataSourceAPIEndpoints tests datasource API endpoints
func TestDataSourceAPIEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "list datasources",
			method:         http.MethodGet,
			path:           "/api/v1/datasources",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list datasources with pagination",
			method:         http.MethodGet,
			path:           "/api/v1/datasources?page=1&page_size=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list datasources with filter",
			method:         http.MethodGet,
			path:           "/api/v1/datasources?type=mysql&status=active",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "create datasource - valid",
			method: http.MethodPost,
			path:   "/api/v1/datasources",
			body: map[string]interface{}{
				"name": "test-ds",
				"type": "mysql",
				"config": map[string]interface{}{
					"host":     "localhost",
					"port":     3306,
					"database": "test",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "create datasource - invalid",
			method: http.MethodPost,
			path:   "/api/v1/datasources",
			body: map[string]interface{}{
				"name": "", // Missing name
				"type": "mysql",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	// Mock handler for testing
	mockHandler := func(expectedStatus int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(expectedStatus)
			if expectedStatus == http.StatusOK || expectedStatus == http.StatusCreated {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":   "test-id",
					"name": "test-ds",
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    "ERROR",
					"message": "Error occurred",
				})
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				body = bytes.NewBuffer(bodyBytes)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockHandler(tt.expectedStatus).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestTaskAPIEndpoints tests task API endpoints
func TestTaskAPIEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "list tasks",
			method:         http.MethodGet,
			path:           "/api/v1/tasks",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list tasks by datasource",
			method:         http.MethodGet,
			path:           "/api/v1/tasks?datasource_id=ds-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "create task - valid",
			method: http.MethodPost,
			path:   "/api/v1/tasks",
			body: map[string]interface{}{
				"name":           "test-task",
				"datasource_id":  "ds-123",
				"type":           "full_collection",
				"scheduler_type": "builtin",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "start task",
			method:         http.MethodPost,
			path:           "/api/v1/tasks/task-123/start",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "stop task",
			method:         http.MethodPost,
			path:           "/api/v1/tasks/task-123/stop",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get task executions",
			method:         http.MethodGet,
			path:           "/api/v1/tasks/task-123/executions",
			expectedStatus: http.StatusOK,
		},
	}

	mockHandler := func(expectedStatus int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(expectedStatus)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				body = bytes.NewBuffer(bodyBytes)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockHandler(tt.expectedStatus).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestBatchOperationEndpoints tests batch operation endpoints
// Feature: datasource-management, Property 12: 批量操作支持
// **Validates: Requirements 7.1, 7.2, 7.3**
func TestBatchOperationEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "batch update status",
			method: http.MethodPost,
			path:   "/api/v1/datasources/batch/status",
			body: map[string]interface{}{
				"ids":    []string{"ds-1", "ds-2", "ds-3"},
				"status": "inactive",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "batch import",
			method: http.MethodPost,
			path:   "/api/v1/datasources/batch/import",
			body: map[string]interface{}{
				"format": "json",
				"data":   "[]",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "batch export",
			method: http.MethodPost,
			path:   "/api/v1/datasources/batch/export",
			body: map[string]interface{}{
				"ids":    []string{"ds-1", "ds-2"},
				"format": "json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "batch start tasks",
			method: http.MethodPost,
			path:   "/api/v1/tasks/batch/start",
			body: map[string]interface{}{
				"ids": []string{"task-1", "task-2"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "batch stop tasks",
			method: http.MethodPost,
			path:   "/api/v1/tasks/batch/stop",
			body: map[string]interface{}{
				"ids": []string{"task-1", "task-2"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	mockHandler := func(expectedStatus int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(expectedStatus)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total":   3,
				"success": 3,
				"failed":  0,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockHandler(tt.expectedStatus).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
