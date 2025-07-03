package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "example/grpc_demo/library"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedUserServiceServer
	pb.UnimplementedLibraryServiceServer
	db *pgxpool.Pool
}

func (s *server) Register(ctx context.Context, user *pb.User) (*pb.AuthResponse, error) {
	username := user.GetUsername()
	password := user.GetPassword()

	// Check if user exists
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username).Scan(&exists)
	if err != nil {
		return &pb.AuthResponse{Message: "Database error"}, err
	}
	if exists {
		return &pb.AuthResponse{Message: "Username already exists"}, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.AuthResponse{Message: "Failed to hash password"}, err
	}

	_, err = s.db.Exec(ctx, "INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, string(hash))
	if err != nil {
		return &pb.AuthResponse{Message: "Failed to create user"}, err
	}

	return &pb.AuthResponse{
		Message: "User registered successfully",
		Token:   "dummy_token",
	}, nil
}

func (s *server) Login(ctx context.Context, creds *pb.UserCredentials) (*pb.AuthResponse, error) {
	username := creds.GetUsername()
	password := creds.GetPassword()

	var hash string
	err := s.db.QueryRow(ctx, "SELECT password_hash FROM users WHERE username=$1", username).Scan(&hash)
	if err != nil {
		return &pb.AuthResponse{Message: "Invalid username or password"}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return &pb.AuthResponse{Message: "Invalid username or password"}, nil
	}

	return &pb.AuthResponse{
		Message: "Login successful",
		Token:   "dummy_token",
	}, nil
}

func (s *server) AddBook(ctx context.Context, book *pb.Book) (*pb.BookResponse, error) {
	if book.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	// Check if book exists
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM books WHERE id=$1)", book.GetId()).Scan(&exists)
	if err != nil {
		return &pb.BookResponse{Id: book.GetId(), Message: "Database error"}, err
	}
	if exists {
		return &pb.BookResponse{Id: book.GetId(), Message: "Book already exists"}, nil
	}
	_, err = s.db.Exec(ctx, "INSERT INTO books (id, title, author) VALUES ($1, $2, $3)", book.GetId(), book.GetTitle(), book.GetAuthor())
	if err != nil {
		return &pb.BookResponse{Id: book.GetId(), Message: "Failed to add book"}, err
	}
	return &pb.BookResponse{Id: book.GetId(), Message: "Book added successfully"}, nil
}

func (s *server) UpdateBook(ctx context.Context, book *pb.Book) (*pb.BookResponse, error) {
	if book.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	res, err := s.db.Exec(ctx, "UPDATE books SET title=$1, author=$2 WHERE id=$3", book.GetTitle(), book.GetAuthor(), book.GetId())
	if err != nil {
		return &pb.BookResponse{Id: book.GetId(), Message: "Failed to update book"}, err
	}
	if res.RowsAffected() == 0 {
		return &pb.BookResponse{Id: book.GetId(), Message: "Book not found"}, nil
	}
	return &pb.BookResponse{Id: book.GetId(), Message: "Book updated successfully"}, nil
}

func (s *server) DeleteBook(ctx context.Context, req *pb.BookRequest) (*pb.BookResponse, error) {
	if req.GetId() == "" {
		return &pb.BookResponse{Id: "", Message: "Book ID is required"}, nil
	}
	res, err := s.db.Exec(ctx, "DELETE FROM books WHERE id=$1", req.GetId())
	if err != nil {
		return &pb.BookResponse{Id: req.GetId(), Message: "Failed to delete book"}, err
	}
	if res.RowsAffected() == 0 {
		return &pb.BookResponse{Id: req.GetId(), Message: "Book not found"}, nil
	}
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
	offset := (page - 1) * pageSize

	rows, err := s.db.Query(ctx, "SELECT id, title, author FROM books ORDER BY id LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var b pb.Book
		if err := rows.Scan(&b.Id, &b.Title, &b.Author); err != nil {
			return nil, err
		}
		books = append(books, &b)
	}
	var totalCount int32
	err = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM books").Scan(&totalCount)
	if err != nil {
		return nil, err
	}
	return &pb.ListBookResponse{Books: books, TotalCount: totalCount}, nil
}

func (s *server) BatchAddBooks(stream pb.LibraryService_BatchAddBooksServer) error {
	var responses []*pb.BookResponse
	ctx := stream.Context()
	for {
		book, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.BatchResponse{Responses: responses})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive book: %v", err)
		}
		if book.GetId() == "" {
			responses = append(responses, &pb.BookResponse{Id: "", Message: "Book ID is required"})
			continue
		}
		var exists bool
		err = s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM books WHERE id=$1)", book.GetId()).Scan(&exists)
		if err != nil {
			responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Database error"})
			continue
		}
		if exists {
			responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Book already exists"})
			continue
		}
		_, err = s.db.Exec(ctx, "INSERT INTO books (id, title, author) VALUES ($1, $2, $3)", book.GetId(), book.GetTitle(), book.GetAuthor())
		if err != nil {
			responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Failed to add book"})
			continue
		}
		responses = append(responses, &pb.BookResponse{Id: book.GetId(), Message: "Book added successfully"})
	}
}

func main() {
	clearDB := flag.Bool("clear-db", false, "Clear all data from database on startup")
	flag.Parse()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	dbpool, err := NewDBPool()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbpool.Close()

	if *clearDB {
		fmt.Println("Clearing database...")
		if err := ClearDatabase(dbpool); err != nil {
			log.Fatalf("failed to clear database: %v", err)
		}
		fmt.Println("Database cleared successfully!")
	}

	if err := RunMigrations(dbpool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{db: dbpool})
	pb.RegisterLibraryServiceServer(s, &server{db: dbpool})

	// Start REST gateway in background
	go StartGateway()

	fmt.Println("gRPC Server is running on port: 50051")
	fmt.Println("REST Gateway is running on port: 8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
