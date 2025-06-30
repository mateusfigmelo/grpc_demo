package main

import (
	"context"
	"fmt"
	"log"

	pb "example/grpc_demo/library"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// Book management tests
	libraryClient := pb.NewLibraryServiceClient(conn)
	book := &pb.Book{
		Id:     "book1",
		Title:  "Go Programming",
		Auther: "John Doe",
	}

	// AddBook
	addResp, err := libraryClient.AddBook(context.Background(), book)
	if err != nil {
		log.Fatalf("could not add book: %v", err)
	}
	fmt.Printf("AddBook Response: %s, ID: %s\n", addResp.GetMessage(), addResp.GetId())

	// UpdateBook
	book.Title = "Advanced Go Programming"
	updateResp, err := libraryClient.UpdateBook(context.Background(), book)
	if err != nil {
		log.Fatalf("could not update book: %v", err)
	}
	fmt.Printf("UpdateBook Response: %s, ID: %s\n", updateResp.GetMessage(), updateResp.GetId())

	// DeleteBook
	deleteResp, err := libraryClient.DeleteBook(context.Background(), &pb.BookRequest{Id: book.GetId()})
	if err != nil {
		log.Fatalf("could not delete book: %v", err)
	}
	fmt.Printf("DeleteBook Response: %s, ID: %s\n", deleteResp.GetMessage(), deleteResp.GetId())
}
