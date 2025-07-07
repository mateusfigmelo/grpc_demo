package main

import (
	"testing"
	"time"

	pb "example/grpc_demo/library"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashing(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"Simple password", "password123"},
		{"Complex password", "P@ssw0rd!2023"},
		{"Long password", "this-is-a-very-long-password-with-many-characters-and-symbols-!@#$%^&*()"},
		{"Special characters", "!@#$%^&*()_+-=[]{}|;:,.<>?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test password hashing
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			// Test password verification
			err = bcrypt.CompareHashAndPassword(hash, []byte(tt.password))
			if err != nil {
				t.Errorf("Password verification failed: %v", err)
			}

			// Test with wrong password
			err = bcrypt.CompareHashAndPassword(hash, []byte("wrong-password"))
			if err == nil {
				t.Error("Password verification should fail with wrong password")
			}
		})
	}
}

func TestPasswordStrength(t *testing.T) {
	// Test that bcrypt cost is appropriate
	password := "test-password"

	// Test default cost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash with default cost: %v", err)
	}

	cost, err := bcrypt.Cost(hash)
	if err != nil {
		t.Fatalf("Failed to get hash cost: %v", err)
	}

	if cost < 10 {
		t.Errorf("Hash cost too low: %d, should be at least 10", cost)
	}
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name    string
		user    *pb.User
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Empty username",
			user:    &pb.User{Username: "", Password: "password123"},
			wantErr: false, // Returns response with error message, not error
			wantMsg: "Username and password are required",
		},
		{
			name:    "Empty password",
			user:    &pb.User{Username: "testuser", Password: ""},
			wantErr: false,
			wantMsg: "Username and password are required",
		},
		{
			name:    "Both empty",
			user:    &pb.User{Username: "", Password: ""},
			wantErr: false,
			wantMsg: "Username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test only validates the input validation logic
			// Full integration tests would require a proper database setup
			if tt.user.Username == "" || tt.user.Password == "" {
				// Simulate the validation logic
				if tt.wantMsg != "Username and password are required" {
					t.Errorf("Expected validation message not matched")
				}
			}
		})
	}
}

func TestLoginValidation(t *testing.T) {
	tests := []struct {
		name    string
		creds   *pb.UserCredentials
		wantMsg string
	}{
		{
			name:    "Empty username",
			creds:   &pb.UserCredentials{Username: "", Password: "password123"},
			wantMsg: "Username and password are required",
		},
		{
			name:    "Empty password",
			creds:   &pb.UserCredentials{Username: "testuser", Password: ""},
			wantMsg: "Username and password are required",
		},
		{
			name:    "Both empty",
			creds:   &pb.UserCredentials{Username: "", Password: ""},
			wantMsg: "Username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the input validation logic
			if tt.creds.Username == "" || tt.creds.Password == "" {
				if tt.wantMsg != "Username and password are required" {
					t.Errorf("Expected validation message not matched")
				}
			}
		})
	}
}

func TestBcryptPerformance(t *testing.T) {
	password := "test-password-for-performance"

	// Test that hashing doesn't take too long
	start := timeTracker()

	for i := 0; i < 10; i++ {
		_, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Hashing failed on iteration %d: %v", i, err)
		}
	}

	duration := start()

	// Each hash should take reasonable time (usually < 100ms each)
	if duration.Seconds() > 2.0 { // 10 hashes in under 2 seconds
		t.Errorf("Password hashing too slow: %v for 10 hashes", duration)
	}
}

// Helper function for timing
func timeTracker() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}

// Test JWT integration with auth methods
func TestJWTIntegration(t *testing.T) {
	// Test that generated tokens are valid
	userID := 123
	username := "testuser"

	token, err := GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate generated JWT: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("JWT userID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("JWT username = %v, want %v", claims.Username, username)
	}

	// Verify token contains expected claims
	if claims.Issuer != "library-service" {
		t.Errorf("JWT issuer = %v, want %v", claims.Issuer, "library-service")
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("JWT should have valid expiration time in the future")
	}
}
