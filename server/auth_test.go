package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key")
	code := m.Run()
	os.Exit(code)
}

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		wantErr  bool
	}{
		{
			name:     "Valid user",
			userID:   1,
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "Another valid user",
			userID:   42,
			username: "anotheruser",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Error("GenerateJWT() returned empty token")
				}

				// Verify the token can be parsed
				claims, err := ValidateJWT(token)
				if err != nil {
					t.Errorf("ValidateJWT() failed to validate generated token: %v", err)
				}

				if claims.UserID != tt.userID {
					t.Errorf("GenerateJWT() userID = %v, want %v", claims.UserID, tt.userID)
				}

				if claims.Username != tt.username {
					t.Errorf("GenerateJWT() username = %v, want %v", claims.Username, tt.username)
				}
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	// Generate a valid token for testing
	validToken, err := GenerateJWT(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate valid token for testing: %v", err)
	}

	// Generate an expired token
	expiredClaims := &Claims{
		UserID:   1,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "library-service",
			Subject:   "1",
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte("test-secret-key"))

	tests := []struct {
		name         string
		token        string
		wantErr      bool
		wantUserID   int
		wantUsername string
	}{
		{
			name:         "Valid token",
			token:        validToken,
			wantErr:      false,
			wantUserID:   1,
			wantUsername: "testuser",
		},
		{
			name:    "Invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "Expired token",
			token:   expiredTokenString,
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Token with wrong secret",
			token:   generateTokenWithWrongSecret(1, "testuser"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateJWT(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims.UserID != tt.wantUserID {
					t.Errorf("ValidateJWT() userID = %v, want %v", claims.UserID, tt.wantUserID)
				}

				if claims.Username != tt.wantUsername {
					t.Errorf("ValidateJWT() username = %v, want %v", claims.Username, tt.wantUsername)
				}
			}
		})
	}
}

func TestExtractTokenFromMetadata(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantErr   bool
		wantToken string
	}{
		{
			name: "Valid Bearer token",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"authorization", "Bearer valid-token-here",
			)),
			wantErr:   false,
			wantToken: "valid-token-here",
		},
		{
			name: "Missing authorization header",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"other-header", "some-value",
			)),
			wantErr: true,
		},
		{
			name: "Invalid authorization format",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"authorization", "Invalid token-here",
			)),
			wantErr: true,
		},
		{
			name: "Empty authorization header",
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				"authorization", "",
			)),
			wantErr: true,
		},
		{
			name:    "No metadata",
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := extractTokenFromMetadata(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractTokenFromMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token != tt.wantToken {
				t.Errorf("extractTokenFromMetadata() token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestGetJWTSecret(t *testing.T) {
	// Test with environment variable set
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test-env-secret")

	secret := getJWTSecret()
	if secret != "test-env-secret" {
		t.Errorf("getJWTSecret() with env var = %v, want %v", secret, "test-env-secret")
	}

	// Test with no environment variable
	os.Unsetenv("JWT_SECRET")
	secret = getJWTSecret()
	if secret == "" {
		t.Error("getJWTSecret() should return default secret when env var is not set")
	}

	// Restore original
	if originalSecret != "" {
		os.Setenv("JWT_SECRET", originalSecret)
	}
}

// Helper function to generate a token with wrong secret for testing
func generateTokenWithWrongSecret(userID int, username string) string {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "library-service",
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret-key"))
	return tokenString
}

// Benchmark tests
func BenchmarkGenerateJWT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateJWT(1, "testuser")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateJWT(b *testing.B) {
	token, err := GenerateJWT(1, "testuser")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateJWT(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}
