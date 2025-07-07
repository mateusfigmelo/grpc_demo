import React, { useState, useEffect } from 'react';
import { Auth } from './components/Auth';
import { BookManager } from './components/BookManager';
import { authAPI } from './services/api';
import './App.css';

function App() {
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check authentication status on app initialization
  useEffect(() => {
    const checkAuthStatus = async () => {
      const isAuthenticated = authAPI.isAuthenticated();
      const currentToken = authAPI.getCurrentToken();
      
      if (isAuthenticated && currentToken) {
        // Also validate with server to ensure user still exists
        const isValidOnServer = await authAPI.validateTokenWithServer();
        if (isValidOnServer) {
          setToken(currentToken);
        } else {
          // Token is invalid on server side (e.g., user deleted, DB cleared)
          authAPI.logout();
          setToken(null);
        }
      } else {
        // Clear any invalid token
        authAPI.logout();
        setToken(null);
      }
      
      setIsLoading(false);
    };

    checkAuthStatus();
  }, []);

  const handleLogin = (newToken: string) => {
    console.log("Login successful", newToken);
    setToken(newToken);
    // Token is automatically stored by the authAPI
  };

  const handleLogout = () => {
    setToken(null);
    // Token is automatically cleared by the authAPI
  };

  // Show loading spinner while checking authentication
  if (isLoading) {
    return (
      <div className="App" style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        flexDirection: 'column',
        gap: '20px'
      }}>
        <div style={{ 
          width: '40px', 
          height: '40px', 
          border: '4px solid #f3f3f3', 
          borderTop: '4px solid #007bff', 
          borderRadius: '50%', 
          animation: 'spin 1s linear infinite' 
        }}></div>
        <p style={{ color: '#6c757d' }}>Checking authentication...</p>
      </div>
    );
  }

  return (
    <div className="App">
      {token ? (
        <BookManager onLogout={handleLogout} />
      ) : (
        <Auth onLogin={handleLogin} />
      )}
    </div>
  );
}

export default App;
