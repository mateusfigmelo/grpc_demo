import React, { useState } from 'react';
import { authAPI, User, AuthResponse, AuthenticationError } from '../services/api';

interface AuthProps {
  onLogin: (token: string) => void;
}

export const Auth: React.FC<AuthProps> = ({ onLogin }) => {
  const [isLogin, setIsLogin] = useState(true);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  const validateForm = (): boolean => {
    if (!username.trim()) {
      setError('Username is required');
      return false;
    }
    if (!password) {
      setError('Password is required');
      return false;
    }
    if (password.length < 6) {
      setError('Password must be at least 6 characters long');
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
      const user: User = { username: username.trim(), password };
      let response: AuthResponse;

      if (isLogin) {
        response = await authAPI.login(user);
        setMessage('Login successful!');
        
        // Call onLogin with the token
        setTimeout(() => {
          onLogin(response.token);
        }, 500);
      } else {
        response = await authAPI.register(user);
        setMessage('Registration successful! You are now logged in.');
        
        // Auto-login after successful registration
        setTimeout(() => {
          onLogin(response.token);
        }, 1000);
      }
    } catch (err) {
      if (err instanceof AuthenticationError) {
        setError('Invalid username or password. Please try again.');
      } else if (err instanceof Error) {
        // Handle specific error messages from the server
        if (err.message.includes('Username already exists')) {
          setError('Username already exists. Please choose a different username or try logging in.');
        } else if (err.message.includes('Invalid username or password')) {
          setError('Invalid username or password. Please check your credentials.');
        } else if (err.message.includes('Username and password are required')) {
          setError('Please provide both username and password.');
        } else {
          setError(err.message);
        }
      } else {
        setError('An unexpected error occurred. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const switchMode = () => {
    setIsLogin(!isLogin);
    setError('');
    setMessage('');
    setUsername('');
    setPassword('');
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2 style={{ textAlign: 'center', marginBottom: '30px', color: '#333' }}>
          {isLogin ? 'Welcome Back' : 'Create Account'}
        </h2>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Username:</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={loading}
              className="form-input"
              placeholder="Enter your username"
              style={{
                opacity: loading ? 0.6 : 1,
                cursor: loading ? 'not-allowed' : 'text'
              }}
            />
          </div>
          
          <div className="form-group">
            <label className="form-label">Password:</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
              className="form-input"
              placeholder="Enter your password"
              style={{
                opacity: loading ? 0.6 : 1,
                cursor: loading ? 'not-allowed' : 'text'
              }}
            />
            {!isLogin && (
              <small style={{ color: '#666', fontSize: '12px' }}>
                Password must be at least 6 characters long
              </small>
            )}
          </div>
          
          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary"
            style={{ 
              width: '100%',
              opacity: loading ? 0.6 : 1,
              cursor: loading ? 'not-allowed' : 'pointer'
            }}
          >
            {loading ? (
              <span>
                {isLogin ? 'Signing In...' : 'Creating Account...'}
              </span>
            ) : (
              isLogin ? 'Sign In' : 'Sign Up'
            )}
          </button>
        </form>
        
        <div style={{ textAlign: 'center', marginTop: '20px' }}>
          <button
            onClick={switchMode}
            disabled={loading}
            className="link-button"
            style={{
              opacity: loading ? 0.6 : 1,
              cursor: loading ? 'not-allowed' : 'pointer'
            }}
          >
            {isLogin ? 'Need to create an account?' : 'Already have an account?'}
          </button>
        </div>
        
        {error && (
          <div className="alert alert-error" style={{ marginTop: '15px' }}>
            <strong>Error:</strong> {error}
          </div>
        )}
        
        {message && (
          <div className="alert alert-success" style={{ marginTop: '15px' }}>
            <strong>Success:</strong> {message}
          </div>
        )}

        {/* Security Notice */}
        <div style={{ 
          marginTop: '20px', 
          padding: '10px', 
          backgroundColor: '#f8f9fa', 
          border: '1px solid #dee2e6', 
          borderRadius: '4px',
          fontSize: '12px',
          color: '#6c757d'
        }}>
          <strong>🔒 Security Notice:</strong> Your password is securely hashed and your session uses JWT tokens that expire in 24 hours.
        </div>
      </div>
    </div>
  );
}; 