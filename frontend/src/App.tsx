import React, { useState } from 'react';
import { Auth } from './components/Auth';
import { BookManager } from './components/BookManager';
import './App.css';

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('authToken'));

  const handleLogin = (newToken: string) => {
    setToken(newToken);
    localStorage.setItem('authToken', newToken);
  };

  const handleLogout = () => {
    setToken(null);
    localStorage.removeItem('authToken');
  };

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
