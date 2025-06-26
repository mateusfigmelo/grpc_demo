package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "example/grpc_demo/library"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUserServiceServer
	users sync.Map
}

func (s *server) Register(ctx context.Context, user *pb.User) (*pb.AuthResponse, error) {
	username := user.GetUsername()
	password := user.GetPassword()

	if _, exists := s.users.Load(username); exists {
		return &pb.AuthResponse{
			Message: "Username already exists",
		}, nil
	}

	s.users.Store(username, password)

	return &pb.AuthResponse{
		Message: "User registered successfully",
		Token:   "dummy_token",
	}, nil
}

func (s *server) Login(ctx context.Context, creds *pb.UserCredentials) (*pb.AuthResponse, error) {
	username := creds.GetUsername()
	password := creds.GetPassword()

	if storedPassword, exists := s.users.Load(username); exists {
		if storedPassword == password {
			return &pb.AuthResponse{
				Message: "Login successful",
				Token:   "dummy_token",
			}, nil
		}
	}

	return &pb.AuthResponse{
		Message: "Invalid username or password",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})

	fmt.Println("Server is running on port: 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
