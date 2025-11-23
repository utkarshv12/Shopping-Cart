import React, { useState, useEffect } from 'react';
import { CartProvider } from './context/CartContext';
import Login from './components/Login';
import ItemsList from './components/ItemsList';
import Header from './components/Header';
import api from './services/api';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import './App.css';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState(null);
  const [userId, setUserId] = useState(null);
  const [theme, setTheme] = useState(() => localStorage.getItem('theme') || 'light');
  const [search, setSearch] = useState('');

  useEffect(() => {
    // Check if user is already logged in
    const savedToken = localStorage.getItem('token');
    const savedUserId = localStorage.getItem('userId');
    
    if (savedToken && savedUserId) {
      setToken(savedToken);
      setUserId(savedUserId);
      setIsLoggedIn(true);
    }
  }, []);

  // apply theme when it changes
  useEffect(() => {
    document.body.classList.toggle('dark', theme === 'dark');
    localStorage.setItem('theme', theme);
  }, [theme]);

  const handleLoginSuccess = (token, userId) => {
    setToken(token);
    setUserId(userId);
    setIsLoggedIn(true);
    localStorage.setItem('token', token);
    localStorage.setItem('userId', userId);
  };

  const handleLogout = () => {
    // Call API logout and clear local storage
    api.logout().finally(() => {
      setIsLoggedIn(false);
      setToken(null);
      setUserId(null);
      localStorage.removeItem('token');
      localStorage.removeItem('userId');
    });
  };

  const toggleTheme = () => {
    const next = theme === 'dark' ? 'light' : 'dark';
    setTheme(next);
    localStorage.setItem('theme', next);
    document.body.classList.toggle('dark', next === 'dark');
  };

  const handleSearch = (q) => setSearch(q || '');

  return (
    <div className="App app-container">
      <ToastContainer position="top-right" autoClose={3000} />
      <CartProvider>
        <Header isLoggedIn={isLoggedIn} onLogout={handleLogout} theme={theme} onToggleTheme={toggleTheme} onSearch={handleSearch} />
        {!isLoggedIn ? (
          <Login onLoginSuccess={handleLoginSuccess} />
        ) : (
          <ItemsList token={token} userId={userId} onLogout={handleLogout} searchQuery={search} />
        )}
      </CartProvider>
      
    </div>
  );
}

export default App;