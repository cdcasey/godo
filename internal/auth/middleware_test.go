package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	validToken, _ := GenerateToken("user-123", "test@example.com", "user", testSecret, time.Hour)
	expiredToken, _ := GenerateToken("user-123", "test@example.com", "user", testSecret, -time.Hour)

	tests := []struct {
		name           string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no bearer prefix",
			authHeader: validToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong prefix",
			authHeader: "Basic " + validToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token.here",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := Middleware(testSecret)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("expected nextCalled=%v, got %v", tt.wantNextCalled, nextCalled)
			}
		})
	}
}
