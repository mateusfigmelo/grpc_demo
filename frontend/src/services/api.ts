// API service for communicating with the gRPC backend via REST endpoints

const API_BASE_URL = 'http://localhost:8080/api/v1';

export interface User {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  message: string;
}

export interface Book {
  id: string;
  title: string;
  author: string;
}

export interface BookResponse {
  id: string;
  message: string;
}

export interface ListBooksResponse {
  books: Book[];
  totalCount: number;
}

// User Authentication
export const authAPI = {
  register: async (user: User): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(user),
    });
    if (!response.ok) {
      throw new Error(`Registration failed: ${response.statusText}`);
    }
    return response.json();
  },

  login: async (credentials: User): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
    });
    if (!response.ok) {
      throw new Error(`Login failed: ${response.statusText}`);
    }
    return response.json();
  },
};

// Book Management
export const bookAPI = {
  addBook: async (book: Book): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(book),
    });
    if (!response.ok) {
      throw new Error(`Add book failed: ${response.statusText}`);
    }
    return response.json();
  },

  updateBook: async (book: Book): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books/${book.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(book),
    });
    if (!response.ok) {
      throw new Error(`Update book failed: ${response.statusText}`);
    }
    return response.json();
  },

  deleteBook: async (id: string): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      throw new Error(`Delete book failed: ${response.statusText}`);
    }
    return response.json();
  },

  listBooks: async (page: number = 1, pageSize: number = 10): Promise<ListBooksResponse> => {
    const response = await fetch(`${API_BASE_URL}/books?page=${page}&page_size=${pageSize}`);
    if (!response.ok) {
      throw new Error(`List books failed: ${response.statusText}`);
    }
    return response.json();
  },
}; 