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

// Token Management
class TokenManager {
  private static instance: TokenManager;
  private token: string | null = null;

  private constructor() {
    // Load token from localStorage on initialization
    this.token = localStorage.getItem('authToken');
  }

  public static getInstance(): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager();
    }
    return TokenManager.instance;
  }

  public setToken(token: string): void {
    this.token = token;
    localStorage.setItem('authToken', token);
    
    // Debug: Log token info
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      const timeUntilExpiry = payload.exp - currentTime;
      console.log('Token set:', {
        exp: payload.exp,
        currentTime,
        timeUntilExpiry,
        expiresAt: new Date(payload.exp * 1000).toLocaleString()
      });
    } catch (error) {
      console.error('Error parsing token:', error);
    }
  }

  public getToken(): string | null {
    return this.token;
  }

  public clearToken(): void {
    this.token = null;
    localStorage.removeItem('authToken');
  }

  public isTokenValid(): boolean {
    if (!this.token) return false;
    
    try {
      // Decode JWT payload (basic validation)
      const payload = JSON.parse(atob(this.token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      
      // Check if token is expired (with small buffer for clock skew)
      // Use smaller buffer (30 seconds) or 10% of token lifetime, whichever is smaller
      const tokenLifetime = payload.exp - payload.iat; // Token lifetime in seconds
      const bufferTime = Math.min(30, Math.floor(tokenLifetime * 0.1)); // 30s or 10% of lifetime
      const isExpired = payload.exp <= (currentTime + bufferTime);
      
      if (isExpired) {
        console.log('Token expired, clearing...', {
          exp: payload.exp,
          currentTime,
          bufferTime,
          timeUntilExpiry: payload.exp - currentTime
        });
        this.clearToken();
        return false;
      }
      
      return true;
    } catch (error) {
      console.error('Invalid token format:', error);
      this.clearToken();
      return false;
    }
  }
}

// Custom error for authentication failures
class AuthenticationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthenticationError';
  }
}

// Helper function to create authenticated headers
const createAuthHeaders = (includeAuth: boolean = true): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  
  if (includeAuth) {
    const token = TokenManager.getInstance().getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }
  
  return headers;
};

// Helper function to handle API responses
const handleResponse = async (response: Response): Promise<any> => {
  if (response.status === 401) {
    // Token expired or invalid - clear it and throw auth error
    TokenManager.getInstance().clearToken();
    throw new AuthenticationError('Authentication failed. Please login again.');
  }
  
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`API Error (${response.status}): ${errorText || response.statusText}`);
  }
  
  const data = await response.json();
  
  // CRITICAL: Check for server-side error messages in successful responses
  // The server may return 200 OK but with an error message for invalid credentials
  if (data.message && !data.token) {
    // If there's a message but no token, it's likely an error
    if (data.message.includes('Invalid username or password') || 
        data.message.includes('Username already exists') || 
        data.message.includes('Username and password are required') ||
        data.message.includes('Failed to')) {
      throw new Error(data.message);
    }
  }
  
  return data;
};

// User Authentication
export const authAPI = {
  register: async (user: User): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: createAuthHeaders(false), // No auth required for registration
      body: JSON.stringify(user),
    });
    
    const data = await handleResponse(response);
    
    // Automatically store token if registration returns one
    if (data.token) {
      TokenManager.getInstance().setToken(data.token);
    }
    
    return data;
  },

  login: async (credentials: User): Promise<AuthResponse> => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: createAuthHeaders(false), // No auth required for login
      body: JSON.stringify(credentials),
    });
    
    const data = await handleResponse(response);
    
    // Store token on successful login
    if (data.token) {
      TokenManager.getInstance().setToken(data.token);
    }
    
    return data;
  },

  logout: (): void => {
    TokenManager.getInstance().clearToken();
  },

  isAuthenticated: (): boolean => {
    const isValid = TokenManager.getInstance().isTokenValid();
    console.log('Auth check:', { isValid });
    return isValid;
  },

  // Validate token with server (checks if user still exists in DB)
  validateTokenWithServer: async (): Promise<boolean> => {
    const token = TokenManager.getInstance().getToken();
    console.log('validateTokenWithServer called, token exists:', !!token);
    if (!token) return false;

    try {
      // Try to make a simple authenticated request to verify token is still valid
      const response = await fetch(`${API_BASE_URL}/books?page=1&page_size=1`, {
        method: 'GET',
        headers: createAuthHeaders(true),
      });

      console.log('Server validation response:', {
        status: response.status,
        ok: response.ok
      });

      if (response.status === 401) {
        // Token is invalid on server side
        console.log('Server rejected token, clearing...');
        TokenManager.getInstance().clearToken();
        return false;
      }

      return response.ok;
    } catch (error) {
      console.error('Token validation failed:', error);
      TokenManager.getInstance().clearToken();
      return false;
    }
  },

  getCurrentToken: (): string | null => {
    return TokenManager.getInstance().getToken();
  }
};

// Book Management (All endpoints require authentication)
export const bookAPI = {
  addBook: async (book: Book): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books`, {
      method: 'POST',
      headers: createAuthHeaders(true), // Auth required
      body: JSON.stringify(book),
    });
    
    return handleResponse(response);
  },

  updateBook: async (book: Book): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books/${book.id}`, {
      method: 'PUT',
      headers: createAuthHeaders(true), // Auth required
      body: JSON.stringify(book),
    });
    
    return handleResponse(response);
  },

  deleteBook: async (id: string): Promise<BookResponse> => {
    const response = await fetch(`${API_BASE_URL}/books/${id}`, {
      method: 'DELETE',
      headers: createAuthHeaders(true), // Auth required
    });
    
    return handleResponse(response);
  },

  listBooks: async (page: number = 1, pageSize: number = 10): Promise<ListBooksResponse> => {
    const response = await fetch(
      `${API_BASE_URL}/books?page=${page}&page_size=${pageSize}`,
      {
        method: 'GET',
        headers: createAuthHeaders(true), // Auth required
      }
    );
    
    return handleResponse(response);
  },
};

// Export token manager for use in components
export { TokenManager, AuthenticationError }; 