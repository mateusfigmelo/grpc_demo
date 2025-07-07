import React, { useState, useEffect } from 'react';
import { bookAPI, Book, AuthenticationError, authAPI } from '../services/api';

interface BookManagerProps {
  onLogout: () => void;
}

export const BookManager: React.FC<BookManagerProps> = ({ onLogout }) => {
  const [books, setBooks] = useState<Book[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(5);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  
  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editingBook, setEditingBook] = useState<Book | null>(null);
  const [formData, setFormData] = useState({ id: '', title: '', author: '' });

  useEffect(() => {
    loadBooks();
  }, [currentPage]);

  // Check if user is still authenticated on component mount
  useEffect(() => {
    const validateAuth = async () => {
      // First check client-side token validity
      if (!authAPI.isAuthenticated()) {
        setError('Your session has expired. Please login again.');
        setTimeout(() => {
          handleLogout();
        }, 2000);
        return;
      }

      // Then validate with server (checks if user exists in DB)
      const isValid = await authAPI.validateTokenWithServer();
      if (!isValid) {
        setError('Your session is no longer valid. Please login again.');
        setTimeout(() => {
          handleLogout();
        }, 2000);
      }
    };

    validateAuth();
  }, []);

  const handleAuthError = (err: Error) => {
    if (err instanceof AuthenticationError) {
      setError('Your session has expired. Please login again.');
      setTimeout(() => {
        handleLogout();
      }, 2000);
    } else {
      setError(err.message);
    }
  };

  const handleLogout = () => {
    authAPI.logout(); // Clear the token
    onLogout();
  };

  const loadBooks = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await bookAPI.listBooks(currentPage, pageSize);
      setBooks(response.books || []);
      setTotalCount(response.totalCount || 0);
    } catch (err) {
      if (err instanceof Error) {
        handleAuthError(err);
      }
    } finally {
      setLoading(false);
    }
  };

  const validateForm = (): boolean => {
    if (!formData.id.trim()) {
      setError('Book ID is required');
      return false;
    }
    if (!formData.title.trim()) {
      setError('Book title is required');
      return false;
    }
    if (!formData.author.trim()) {
      setError('Book author is required');
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setMessage('');

    if (!validateForm()) {
      setLoading(false);
      return;
    }

    try {
      const bookData = {
        id: formData.id.trim(),
        title: formData.title.trim(),
        author: formData.author.trim()
      };

      if (editingBook) {
        await bookAPI.updateBook(bookData);
        setMessage('Book updated successfully!');
      } else {
        await bookAPI.addBook(bookData);
        setMessage('Book added successfully!');
      }
      
      setFormData({ id: '', title: '', author: '' });
      setEditingBook(null);
      setShowForm(false);
      loadBooks();
    } catch (err) {
      if (err instanceof Error) {
        handleAuthError(err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (book: Book) => {
    setFormData(book);
    setEditingBook(book);
    setShowForm(true);
    setError('');
    setMessage('');
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this book?')) return;
    
    setLoading(true);
    setError('');
    setMessage('');
    
    try {
      await bookAPI.deleteBook(id);
      setMessage('Book deleted successfully!');
      loadBooks();
    } catch (err) {
      if (err instanceof Error) {
        handleAuthError(err);
      }
    } finally {
      setLoading(false);
    }
  };

  const cancelForm = () => {
    setShowForm(false);
    setFormData({ id: '', title: '', author: '' });
    setEditingBook(null);
    setError('');
    setMessage('');
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
        <h1>Library Management System</h1>
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <span style={{ fontSize: '14px', color: '#6c757d' }}>
            🔒 Authenticated Session
          </span>
          <button 
            onClick={handleLogout} 
            style={{ 
              padding: '8px 16px', 
              backgroundColor: '#dc3545', 
              color: 'white', 
              border: 'none', 
              borderRadius: '4px', 
              cursor: 'pointer',
              fontSize: '14px'
            }}
          >
            Logout
          </button>
        </div>
      </div>

      {/* Add Book Button */}
      <div style={{ marginBottom: '20px' }}>
        <button
          onClick={() => {
            setFormData({ id: '', title: '', author: '' });
            setEditingBook(null);
            setShowForm(true);
            setError('');
            setMessage('');
          }}
          disabled={loading}
          style={{ 
            padding: '10px 20px', 
            backgroundColor: '#28a745', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px', 
            cursor: loading ? 'not-allowed' : 'pointer',
            opacity: loading ? 0.6 : 1
          }}
        >
          Add New Book
        </button>
      </div>

      {/* Form Modal */}
      {showForm && (
        <div style={{ backgroundColor: '#f8f9fa', padding: '20px', borderRadius: '8px', marginBottom: '20px', border: '1px solid #dee2e6' }}>
          <h3>{editingBook ? 'Edit Book' : 'Add New Book'}</h3>
          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Book ID:</label>
              <input
                type="text"
                value={formData.id}
                onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                required
                disabled={!!editingBook || loading}
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ced4da', 
                  borderRadius: '4px', 
                  boxSizing: 'border-box',
                  opacity: (!!editingBook || loading) ? 0.6 : 1,
                  cursor: (!!editingBook || loading) ? 'not-allowed' : 'text'
                }}
                placeholder="Enter unique book ID"
              />
              {!editingBook && (
                <small style={{ color: '#666', fontSize: '12px' }}>
                  Book ID must be unique and cannot be changed after creation
                </small>
              )}
            </div>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Title:</label>
              <input
                type="text"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                required
                disabled={loading}
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ced4da', 
                  borderRadius: '4px', 
                  boxSizing: 'border-box',
                  opacity: loading ? 0.6 : 1,
                  cursor: loading ? 'not-allowed' : 'text'
                }}
                placeholder="Enter book title"
              />
            </div>
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Author:</label>
              <input
                type="text"
                value={formData.author}
                onChange={(e) => setFormData({ ...formData, author: e.target.value })}
                required
                disabled={loading}
                style={{ 
                  width: '100%', 
                  padding: '8px', 
                  border: '1px solid #ced4da', 
                  borderRadius: '4px', 
                  boxSizing: 'border-box',
                  opacity: loading ? 0.6 : 1,
                  cursor: loading ? 'not-allowed' : 'text'
                }}
                placeholder="Enter author name"
              />
            </div>
            <div style={{ display: 'flex', gap: '10px' }}>
              <button 
                type="submit" 
                disabled={loading} 
                style={{ 
                  padding: '8px 16px', 
                  backgroundColor: '#007bff', 
                  color: 'white', 
                  border: 'none', 
                  borderRadius: '4px', 
                  cursor: loading ? 'not-allowed' : 'pointer',
                  opacity: loading ? 0.6 : 1
                }}
              >
                {loading ? 'Saving...' : (editingBook ? 'Update' : 'Add')}
              </button>
              <button 
                type="button" 
                onClick={cancelForm}
                disabled={loading}
                style={{ 
                  padding: '8px 16px', 
                  backgroundColor: '#6c757d', 
                  color: 'white', 
                  border: 'none', 
                  borderRadius: '4px', 
                  cursor: loading ? 'not-allowed' : 'pointer',
                  opacity: loading ? 0.6 : 1
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Messages */}
      {error && (
        <div style={{ marginBottom: '20px', padding: '10px', backgroundColor: '#f8d7da', color: '#721c24', border: '1px solid #f5c6cb', borderRadius: '4px' }}>
          <strong>Error:</strong> {error}
        </div>
      )}
      {message && (
        <div style={{ marginBottom: '20px', padding: '10px', backgroundColor: '#d4edda', color: '#155724', border: '1px solid #c3e6cb', borderRadius: '4px' }}>
          <strong>Success:</strong> {message}
        </div>
      )}

      {/* Books Table */}
      <div style={{ backgroundColor: 'white', borderRadius: '8px', overflow: 'hidden', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead style={{ backgroundColor: '#f8f9fa' }}>
            <tr>
              <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>ID</th>
              <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>Title</th>
              <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>Author</th>
              <th style={{ padding: '12px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={4} style={{ padding: '20px', textAlign: 'center' }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '10px' }}>
                    <div style={{ 
                      width: '20px', 
                      height: '20px', 
                      border: '2px solid #f3f3f3', 
                      borderTop: '2px solid #007bff', 
                      borderRadius: '50%', 
                      animation: 'spin 1s linear infinite' 
                    }}></div>
                    Loading books...
                  </div>
                </td>
              </tr>
            ) : books.length === 0 ? (
              <tr>
                <td colSpan={4} style={{ padding: '20px', textAlign: 'center', color: '#6c757d' }}>
                  No books found. Click "Add New Book" to get started.
                </td>
              </tr>
            ) : (
              books.map((book) => (
                <tr key={book.id}>
                  <td style={{ padding: '12px', borderBottom: '1px solid #dee2e6' }}>{book.id}</td>
                  <td style={{ padding: '12px', borderBottom: '1px solid #dee2e6' }}>{book.title}</td>
                  <td style={{ padding: '12px', borderBottom: '1px solid #dee2e6' }}>{book.author}</td>
                  <td style={{ padding: '12px', borderBottom: '1px solid #dee2e6' }}>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button
                        onClick={() => handleEdit(book)}
                        disabled={loading}
                        style={{
                          padding: '4px 8px',
                          backgroundColor: '#ffc107',
                          color: '#212529',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: loading ? 'not-allowed' : 'pointer',
                          fontSize: '12px',
                          opacity: loading ? 0.6 : 1
                        }}
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(book.id)}
                        disabled={loading}
                        style={{
                          padding: '4px 8px',
                          backgroundColor: '#dc3545',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: loading ? 'not-allowed' : 'pointer',
                          fontSize: '12px',
                          opacity: loading ? 0.6 : 1
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', marginTop: '20px', gap: '10px' }}>
          <button
            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
            disabled={currentPage === 1 || loading}
            style={{
              padding: '8px 12px',
              border: '1px solid #dee2e6',
              backgroundColor: currentPage === 1 ? '#f8f9fa' : 'white',
              cursor: (currentPage === 1 || loading) ? 'not-allowed' : 'pointer',
              borderRadius: '4px'
            }}
          >
            Previous
          </button>
          
          <span style={{ padding: '8px 16px', color: '#6c757d' }}>
            Page {currentPage} of {totalPages} ({totalCount} total books)
          </span>
          
          <button
            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages || loading}
            style={{
              padding: '8px 12px',
              border: '1px solid #dee2e6',
              backgroundColor: currentPage === totalPages ? '#f8f9fa' : 'white',
              cursor: (currentPage === totalPages || loading) ? 'not-allowed' : 'pointer',
              borderRadius: '4px'
            }}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}; 