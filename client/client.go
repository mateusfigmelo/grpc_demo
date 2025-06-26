package main

import (
	"context"
	"fmt"
	"log"

	pb "example/grpc_demo/library"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	username := "testUser"
	password := "password123"

	regResp, err := client.Register(context.Background(), &pb.User{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	fmt.Printf("Register Response: %s, Token: %s\n", regResp.GetMessage(), regResp.GetToken())

	loginResp, err := client.Login(context.Background(), &pb.UserCredentials{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	fmt.Printf("Login Response: %s, Token: %s\n", loginResp.GetMessage(), loginResp.GetToken())
}
