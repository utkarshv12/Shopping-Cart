import React from 'react';
import { useCart } from '../context/CartContext';
import './Header.css';

function Header({ isLoggedIn, onLogout, theme, onToggleTheme, onSearch }) {
  const { count, loading } = useCart() || { count: 0, loading: false };

  return (
    <header className="app-header">
      <div className="header-left">
        <div className="logo">🛍️ Shoply</div>
        <div className="search">
          <input
            placeholder="Search products..."
            onChange={(e) => onSearch && onSearch(e.target.value)}
            aria-label="Search products"
          />
        </div>
      </div>

      <div className="header-right">
        <button className="icon-btn" onClick={onToggleTheme} title="Toggle theme">
          {theme === 'dark' ? '🌙' : '☀️'}
        </button>
        {isLoggedIn && (
          <>
            <button className="icon-btn cart-btn" title="Open cart">
              🛒
              <span className={`cart-badge ${loading ? 'loading' : ''}`}>
                {count > 99 ? '99+' : count}
              </span>
            </button>
            <button className="btn-logout small" onClick={onLogout}>
              Logout
            </button>
          </>
        )}
      </div>
    </header>
  );
}

export default Header;
