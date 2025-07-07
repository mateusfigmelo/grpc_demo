package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWT secret key - in production, this should be loaded from environment variables
var jwtSecret = []byte(getJWTSecret())

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// getJWTSecret returns the JWT secret from environment or default
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secure-secret-key-change-this-in-production"
	}
	return secret
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID int, username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "library-service",
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// extractTokenFromMetadata extracts the token from gRPC metadata
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata found")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", errors.New("no authorization header")
	}

	// Expected format: "Bearer <token>"
	if len(authHeader[0]) < 7 || authHeader[0][:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[0][7:], nil
}

// validateUserExistsInDB validates that the user from JWT claims still exists in the database
func validateUserExistsInDB(ctx context.Context, db *pgxpool.Pool, userID int, username string) error {
	var dbUserID int
	var dbUsername string

	err := db.QueryRow(ctx, "SELECT id, username FROM users WHERE id=$1 AND username=$2", userID, username).Scan(&dbUserID, &dbUsername)
	if err != nil {
		return fmt.Errorf("user not found in database")
	}

	// Double-check that the data matches exactly
	if dbUserID != userID || dbUsername != username {
		return fmt.Errorf("user data mismatch")
	}

	return nil
}

// CreateAuthInterceptor creates a gRPC unary interceptor for authentication with database access
func CreateAuthInterceptor(db *pgxpool.Pool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for Register and Login methods
		if info.FullMethod == "/library.UserService/Register" || info.FullMethod == "/library.UserService/Login" {
			return handler(ctx, req)
		}

		token, err := extractTokenFromMetadata(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		claims, err := ValidateJWT(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// CRITICAL: Validate that the user still exists in the database
		err = validateUserExistsInDB(ctx, db, claims.UserID, claims.Username)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "user validation failed: %v", err)
		}

		// Add user info to context for use in handlers
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)

		return handler(ctx, req)
	}
}

// CreateStreamAuthInterceptor creates a gRPC stream interceptor for authentication with database access
func CreateStreamAuthInterceptor(db *pgxpool.Pool) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for methods that don't require it
		// For now, all streaming methods require authentication

		token, err := extractTokenFromMetadata(ss.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		claims, err := ValidateJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// CRITICAL: Validate that the user still exists in the database
		err = validateUserExistsInDB(ss.Context(), db, claims.UserID, claims.Username)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "user validation failed: %v", err)
		}

		// Create a new context with user info
		ctx := context.WithValue(ss.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)

		// Wrap the stream with the new context
		wrappedStream := &contextServerStream{ss, ctx}

		return handler(srv, wrappedStream)
	}
}

// contextServerStream wraps grpc.ServerStream to override context
type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *contextServerStream) Context() context.Context {
	return w.ctx
}
