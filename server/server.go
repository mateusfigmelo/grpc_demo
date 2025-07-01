package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "example/grpc_demo/library"

	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedUserServiceServer
	pb.UnimplementedLibraryServiceServer
	users sync.Map
	books sync.Map // key: book id, value: *pb.Book
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

func (s *server) AddBook(ctx context.Context, book *pb.Book) (*pb.BookResponse, error) {
	if book.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	if _, exists := s.books.Load(book.GetId()); exists {
		return &pb.BookResponse{Id: book.GetId(), Message: "Book already exists"}, nil
	}
	s.books.Store(book.GetId(), book)
	return &pb.BookResponse{Id: book.GetId(), Message: "Book added successfully"}, nil
}

func (s *server) UpdateBook(ctx context.Context, book *pb.Book) (*pb.BookResponse, error) {
	if book.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	if _, exists := s.books.Load(book.GetId()); !exists {
		return &pb.BookResponse{Id: book.GetId(), Message: "Book not found"}, nil
	}
	s.books.Store(book.GetId(), book)
	return &pb.BookResponse{Id: book.GetId(), Message: "Book updated successfully"}, nil
}

func (s *server) DeleteBook(ctx context.Context, req *pb.BookRequest) (*pb.BookResponse, error) {
	if req.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	if _, exists := s.books.Load(req.GetId()); !exists {
		return &pb.BookResponse{Id: req.GetId(), Message: "Book not found"}, nil
	}
	s.books.Delete(req.GetId())
	return &pb.BookResponse{Id: req.GetId(), Message: "Book deleted successfully"}, nil
}

func (s *server) ListBooks(ctx context.Context, req *pb.ListBookRequest) (*pb.ListBookResponse, error) {
	page := req.GetPage()
	pageSize := req.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize

	var (
		books      []*pb.Book
		totalCount int32
		idx        int32
	)

	s.books.Range(func(_, value any) bool {
		if book, ok := value.(*pb.Book); ok {
			if idx >= start && int32(len(books)) < pageSize {
				books = append(books, book)
			}
			idx++
		}
		return true
	})
	totalCount = idx

	return &pb.ListBookResponse{
		Books:      books,
		TotalCount: totalCount,
	}, nil
}

func (s *server) BatchAddBooks(stream pb.LibraryService_BatchAddBooksServer) error {
	var responses []*pb.BookResponse
	for {
		book, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.BatchResponse{Reponses: responses})
		}
		if err != nil {
			return status.Errorf(13, "failed to receive book: %v", err)
		}

		if book.GetId() == "" {
			responses = append(responses, &pb.BookResponse{Id: "", Message: "Book ID is required"})
			continue
		}
		if _, exists := s.books.Load(book.GetId()); exists {
			responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Book already exists"})
			continue
		}
		s.books.Store(book.GetId(), book)
		responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Book added successfully"})
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})
	pb.RegisterLibraryServiceServer(s, &server{})

	fmt.Println("Server is running on port: 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
