# Library Management System

A full-stack application with a gRPC backend using PostgreSQL and a modern React frontend.

## Project Structure

```
grpc_demo/
├── server/                 # Go gRPC backend
│   ├── server.go          # Main server with gRPC services
│   ├── db.go              # Database connection and helpers
│   ├── gateway.go         # REST gateway for frontend
│   └── migrations.sql     # Database schema
├── frontend/              # React TypeScript frontend
│   ├── src/
│   │   ├── components/    # React components (Auth, BookManager)
│   │   ├── services/      # API service layer
│   │   └── App.tsx        # Main app component
├── library/               # Protocol buffer definitions
│   ├── library.proto      # gRPC service definitions
│   ├── buf.yaml          # Buf configuration
│   ├── library.pb.go     # Generated protobuf messages
│   ├── library_grpc.pb.go # Generated gRPC service code
│   └── library.pb.gw.go  # Generated gateway code
├── client/                # CLI client for testing
├── buf.gen.yaml          # Buf generation config
├── buf.work.yaml         # Buf workspace config
├── generate-buf.sh       # Script to generate protobuf files
└── .env                  # Environment variables (git ignored)
```

## Prerequisites

- **Go 1.21+**
- **Node.js 18+**
- **PostgreSQL 12+**
- **Buf CLI** (for protobuf generation)

## Setup Instructions

### 1. Database Setup

1. Install and start PostgreSQL
2. Create a database for the application:
   ```sql
   CREATE DATABASE library_db;
   ```

### 2. Environment Configuration

The `.env` file should already exist in the root directory with default PostgreSQL settings:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=library_db
```

Edit these values according to your PostgreSQL setup.

### 3. Protocol Buffer Generation

Generate the protobuf files using Buf:
```bash
bash generate-buf.sh
```

### 4. Backend Setup

1. Install Go dependencies:
   ```bash
   go mod download
   ```

2. Start the gRPC server (from the server directory):
   ```bash
   cd server
   go run .
   ```

   Optional flags:
   - `--clear-db`: Drop and recreate all database tables
   
   Example:
   ```bash
   go run . --clear-db
   ```

   This will start:
   - gRPC server on port `50051`
   - REST gateway on port `8080`
   - Automatic database migrations

### 5. Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm start
   ```

   This will start the React app on `http://localhost:3000`

## Usage

### Web Interface

1. Open `http://localhost:3000` in your browser
2. Register a new user or login with existing credentials
3. Manage books using the web interface:
   - Add new books
   - Edit existing books
   - Delete books
   - View paginated book lists

### API Endpoints

The REST gateway exposes the following endpoints:

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/books` - List books (with pagination)
- `POST /api/v1/books` - Add a new book
- `PUT /api/v1/books/{id}` - Update a book
- `DELETE /api/v1/books/{id}` - Delete a book

### gRPC Services

Direct gRPC access is available on `localhost:50051`:

- **UserService**: Register, Login
- **LibraryService**: AddBook, UpdateBook, DeleteBook, ListBooks, BatchAddBooks

### CLI Client

Test the gRPC services directly:
```bash
cd client
go run client.go
```

## Features

### Backend
- ✅ gRPC server with PostgreSQL integration
- ✅ User authentication with bcrypt password hashing
- ✅ CRUD operations for books
- ✅ Batch book operations (streaming)
- ✅ Pagination support
- ✅ Automatic database migrations
- ✅ Database clearing functionality
- ✅ REST gateway for frontend communication
- ✅ CORS support for cross-origin requests

### Frontend
- ✅ Modern React with TypeScript
- ✅ User authentication (login/register)
- ✅ Book management interface
- ✅ Pagination for book lists
- ✅ Error handling and user feedback
- ✅ Responsive design with modern UI
- ✅ Authentication state management

### Development Tools
- ✅ Buf for protobuf management
- ✅ Automated protobuf generation
- ✅ CLI client for testing
- ✅ Database reset functionality

## Development

### Regenerating Protocol Buffers

When you modify `library/library.proto`, regenerate the files:
```bash
bash generate-buf.sh
```

### Database Management

Clear and recreate database tables:
```bash
cd server
go run . --clear-db
```

### Running Tests

Backend:
```bash
cd client
go run client.go  # Test all gRPC services
```

Frontend:
```bash
cd frontend
npm test
```

### Building for Production

Backend:
```bash
cd server
go build -o ../bin/server .
```

Frontend:
```bash
cd frontend
npm run build
```

## Troubleshooting

### Common Issues

1. **Database connection failed**
   - Ensure PostgreSQL is running
   - Verify database credentials in `.env`
   - Check if the database exists
   - Try using `--clear-db` flag to reset

2. **Frontend API calls failing**
   - Ensure the backend server is running on port 8080
   - Check CORS configuration
   - Verify API endpoints are accessible
   - Check browser console for errors

3. **gRPC errors**
   - Ensure the gRPC server is running on port 50051
   - Regenerate protobuf files with `bash generate-buf.sh`
   - Check if both servers started successfully

4. **Protobuf compilation errors**
   - Install Buf CLI
   - Check `buf.yaml` configuration
   - Ensure googleapis dependencies are available

### Environment Variables

Required environment variables for the backend:

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database username (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: library_db)

## Architecture

### Backend Architecture
- **gRPC Services**: Core business logic with UserService and LibraryService
- **PostgreSQL**: Persistent data storage with automatic migrations
- **REST Gateway**: HTTP/JSON API for frontend using grpc-gateway
- **Buf**: Modern protobuf management and generation

### Frontend Architecture
- **React Components**: Auth and BookManager components
- **API Service Layer**: HTTP client for backend communication
- **TypeScript**: Full type safety
- **Modern CSS**: Responsive design with gradients and animations

## Technology Stack

### Backend
- Go 1.21+
- gRPC with Protocol Buffers
- PostgreSQL with pgx driver
- grpc-gateway for REST API
- bcrypt for password hashing
- Buf for protobuf management

### Frontend
- React 18 with TypeScript
- Modern CSS with responsive design
- Fetch API for HTTP requests
- Local storage for state persistence

## Recent Updates

- ✅ Fixed gRPC protocol errors by simplifying gateway configuration
- ✅ Added database clearing functionality with `--clear-db` flag
- ✅ Implemented proper error handling and logging
- ✅ Updated to modern gRPC client patterns
- ✅ Enhanced UI with modern design patterns
- ✅ Restored BatchAddBooks functionality with correct interface 