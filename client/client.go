package main

import (
	"context"
	"fmt"
	"log"

	pb "example/grpc_demo/library"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AuthenticatedClient wraps the gRPC client with authentication
type AuthenticatedClient struct {
	token string
}

// addAuthToContext adds the authentication token to the context
func (a *AuthenticatedClient) addAuthToContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+a.token)
}

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	username := "testUser"
	password := "password123"

	// Registration (no authentication required)
	regResp, err := client.Register(context.Background(), &pb.User{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	fmt.Printf("Register Response: %s, Token: %s\n", regResp.GetMessage(), regResp.GetToken())

	// Login (no authentication required)
	loginResp, err := client.Login(context.Background(), &pb.UserCredentials{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	fmt.Printf("Login Response: %s, Token: %s\n", loginResp.GetMessage(), loginResp.GetToken())

	// Create authenticated client wrapper
	authClient := &AuthenticatedClient{token: loginResp.GetToken()}

	// Book management tests (all require authentication)
	libraryClient := pb.NewLibraryServiceClient(conn)
	book := &pb.Book{
		Id:     "book1",
		Title:  "Go Programming",
		Author: "John Doe",
	}

	// AddBook (with authentication)
	addResp, err := libraryClient.AddBook(authClient.addAuthToContext(context.Background()), book)
	if err != nil {
		log.Fatalf("could not add book: %v", err)
	}
	fmt.Printf("AddBook Response: %s, ID: %s\n", addResp.GetMessage(), addResp.GetId())

	// UpdateBook (with authentication)
	book.Title = "Advanced Go Programming"
	updateResp, err := libraryClient.UpdateBook(authClient.addAuthToContext(context.Background()), book)
	if err != nil {
		log.Fatalf("could not update book: %v", err)
	}
	fmt.Printf("UpdateBook Response: %s, ID: %s\n", updateResp.GetMessage(), updateResp.GetId())

	// DeleteBook (with authentication)
	deleteResp, err := libraryClient.DeleteBook(authClient.addAuthToContext(context.Background()), &pb.BookRequest{Id: book.GetId()})
	if err != nil {
		log.Fatalf("could not delete book: %v", err)
	}
	fmt.Printf("DeleteBook Response: %s, ID: %s\n", deleteResp.GetMessage(), deleteResp.GetId())

	// Add multiple books for pagination test (with authentication)
	for i := 1; i <= 10; i++ {
		b := &pb.Book{
			Id:     fmt.Sprintf("book%d", i),
			Title:  fmt.Sprintf("Book Title %d", i),
			Author: fmt.Sprintf("Author %d", i),
		}
		_, err := libraryClient.AddBook(authClient.addAuthToContext(context.Background()), b)
		if err != nil {
			log.Printf("could not add book %d: %v", i, err)
		}
	}

	// ListBooks (with authentication)
	listResp, err := libraryClient.ListBooks(authClient.addAuthToContext(context.Background()), &pb.ListBookRequest{
		Page:     1,
		PageSize: 5,
	})
	if err != nil {
		log.Fatalf("could not list books: %v", err)
	}
	fmt.Printf("ListBooks Response: total=%d\n", listResp.GetTotalCount())
	for i, b := range listResp.GetBooks() {
		fmt.Printf("Book %d: ID=%s, Title=%s, Author=%s\n", i+1, b.GetId(), b.GetTitle(), b.GetAuthor())
	}

	// BatchAddBooks (client-side streaming with authentication)
	batchStream, err := libraryClient.BatchAddBooks(authClient.addAuthToContext(context.Background()))
	if err != nil {
		log.Fatalf("could not start batch add books: %v", err)
	}
	for i := 11; i <= 15; i++ {
		b := &pb.Book{
			Id:     fmt.Sprintf("batchbook%d", i),
			Title:  fmt.Sprintf("Batch Book Title %d", i),
			Author: fmt.Sprintf("Batch Author %d", i),
		}
		if err := batchStream.Send(b); err != nil {
			log.Fatalf("failed to send book %d: %v", i, err)
		}
	}
	batchResp, err := batchStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive batch response: %v", err)
	}
	fmt.Println("BatchAddBooks Response:")
	for i, r := range batchResp.GetResponses() {
		fmt.Printf("  Book %d: ID=%s, Message=%s\n", i+1, r.GetId(), r.GetMessage())
	}

	fmt.Println("All tests completed successfully!")

	// Test unauthorized access (optional - to demonstrate authentication works)
	fmt.Println("\n--- Testing unauthorized access ---")
	_, err = libraryClient.AddBook(context.Background(), &pb.Book{
		Id:     "unauthorized-book",
		Title:  "This should fail",
		Author: "Anonymous",
	})
	if err != nil {
		fmt.Printf("Expected error for unauthorized request: %v\n", err)
	} else {
		fmt.Println("WARNING: Unauthorized request succeeded (this should not happen)")
	}
}
