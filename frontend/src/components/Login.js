import React, { useState } from 'react';
import api from '../services/api';
import './Login.css';

function Login({ onLoginSuccess }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isSignup, setIsSignup] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      if (isSignup) {
        // Signup
        await api.signup(username, password);
        alert('Signup successful! Please login.');
        setIsSignup(false);
        setPassword('');
      } else {
        // Login
        const response = await api.login(username, password);
        if (response.data.token) {
          // Save token to localStorage so api helper picks it up
          localStorage.setItem('token', response.data.token);
          onLoginSuccess(response.data.token, response.data.user_id);
        }
      }
    } catch (error) {
      if (error.response && error.response.status === 401) {
        window.alert('Invalid username/password');
      } else {
        window.alert(error.response?.data?.error || 'An error occurred');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-box">
        <h1>{isSignup ? 'Sign Up' : 'Login'}</h1>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="Enter username"
            />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="Enter password"
            />
          </div>
          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Please wait...' : (isSignup ? 'Sign Up' : 'Login')}
          </button>
        </form>
        <p className="toggle-link">
          {isSignup ? 'Already have an account?' : "Don't have an account?"}{' '}
          <span onClick={() => setIsSignup(!isSignup)}>
            {isSignup ? 'Login' : 'Sign Up'}
          </span>
        </p>
      </div>
    </div>
  );
}

export default Login;