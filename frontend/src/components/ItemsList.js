import React, { useState, useEffect } from 'react';
import api from '../services/api';
import Cart from './Cart';
import Orders from './Orders';
import AddItemModal from './AddItemModal';
import ProductModal from './ProductModal';
import { toast } from 'react-toastify';
import './ItemsList.css';

function ItemsList({ token, userId, onLogout, searchQuery = '' }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchItems();
  }, []);

  const fetchItems = async () => {
    try {
      const response = await api.getItems();
      setItems(response.data.items || []);
    } catch (error) {
      toast.error('Failed to load items');
    } finally {
      setLoading(false);
    }
  };

  const [showCart, setShowCart] = useState(false);
  const [showOrders, setShowOrders] = useState(false);
  const [showAddItem, setShowAddItem] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);

  const addToCart = async (itemId) => {
    try {
      await api.addToCart(itemId);
      toast.success('Item added to cart!');
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to add item to cart');
    }
  };

  const viewCart = async () => {
    // open cart modal
    setShowCart(true);
  };

  const viewOrderHistory = async () => {
    // show orders modal
    setShowOrders(true);
  };

  const checkout = async () => {
    try {
      const response = await api.createOrder();
      toast.success(`Order successful! Order ID: ${response.data.order.id}`);
      // Refresh items and close cart if open
      fetchItems();
      setShowCart(false);
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to create order');
    }
  };

  if (loading) {
    // show skeleton grid while loading
    return (
      <div className="items-container">
        <header className="items-header">
          <h1>Shopping Portal</h1>
          <div className="header-actions">
            <div className="skeleton skeleton-button" />
            <div className="skeleton skeleton-button" />
            <div className="skeleton skeleton-button" />
          </div>
        </header>
        <div className="items-grid">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="item-card skeleton-card">
              <div className="skeleton skeleton-image" />
              <div style={{ padding: 12 }}>
                <div className="skeleton skeleton-line short" />
                <div className="skeleton skeleton-line" />
                <div className="skeleton skeleton-line tiny" />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  const q = (searchQuery || '').toLowerCase().trim();
  const displayedItems = q
    ? items.filter((it) => (it.name || '').toLowerCase().includes(q) || (it.description || '').toLowerCase().includes(q))
    : items;

  return (
    <div className="items-container">
      <header className="items-header">
        <h1>Shopping Portal</h1>
        <div className="header-actions">
          <button onClick={fetchItems} className="btn-secondary">⟳ Refresh</button>
          <button onClick={viewCart} className="btn-secondary">🛒 Cart</button>
          <button onClick={() => setShowAddItem(true)} className="btn-secondary">＋ Add Item</button>
          <button onClick={viewOrderHistory} className="btn-secondary">📦 Order History</button>
          <button onClick={checkout} className="btn-primary">✓ Checkout</button>
          <button onClick={onLogout} className="btn-logout">Logout</button>
        </div>
      </header>

      {showCart && (
        <Cart
          onClose={() => setShowCart(false)}
          onCheckout={() => {
            // refresh items and close cart
            fetchItems();
            setShowCart(false);
          }}
        />
      )}

      {showOrders && <Orders onClose={() => setShowOrders(false)} />}

      {showAddItem && (
        <AddItemModal
          onClose={() => setShowAddItem(false)}
          onCreated={() => {
            fetchItems();
            setShowAddItem(false);
          }}
        />
      )}

      <div className="items-grid">
        {displayedItems.length === 0 ? (
          <p className="no-items">No items available</p>
        ) : (
          displayedItems.map((item) => (
            <div key={item.id} className="item-card" onClick={() => setSelectedItem(item)}>
              <img className="item-image" src={item.image_url ? `${process.env.REACT_APP_API_BASE || 'http://localhost:8080'}${item.image_url}` : `https://picsum.photos/seed/${item.id}/600/360`} alt={item.name} />
              <div className="item-info">
                <h3>{item.name}</h3>
                <p className="item-description">{item.description}</p>
                <p className="item-price">${item.price.toFixed(2)}</p>
              </div>
              <div className="item-actions">
                <button className="btn-add" onClick={(e) => { e.stopPropagation(); addToCart(item.id); }}>Add to Cart</button>
                <button className="btn-danger" onClick={async (e) => { e.stopPropagation();
                  if (!window.confirm('Delete this item?')) return; try { await api.deleteItem(item.id); toast.success('Item deleted'); fetchItems(); } catch(err){ toast.error(err.response?.data?.error || 'Failed to delete item'); } }}>Delete</button>
              </div>
            </div>
          ))
        )}
      </div>
      {selectedItem && (
        <ProductModal item={selectedItem} onClose={() => setSelectedItem(null)} />
      )}
    </div>
  );
}

export default ItemsList;