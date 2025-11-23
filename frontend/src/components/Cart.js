import React, { useEffect, useState } from 'react';
import { toast } from 'react-toastify';
import api from '../services/api';
import './Cart.css';

function Cart({ onClose, onCheckout }) {
  const [cart, setCart] = useState(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const loadCart = async () => {
    setLoading(true);
    try {
      const res = await api.getUserCart();
      setCart(res.data.cart);
    } catch (err) {
      setCart(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCart();
  }, []);

  const removeItem = async (itemId) => {
    setBusy(true);
    try {
      await api.removeFromCart(itemId);
      await loadCart();
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to remove item');
    } finally {
      setBusy(false);
    }
  };

  const total = cart?.items?.reduce((s, it) => s + (it.item?.price || 0) * it.quantity, 0) || 0;

  return (
    <div className="cart-overlay" onClick={onClose}>
      <div className="cart-drawer glass fade-in" onClick={(e) => e.stopPropagation()}>
        <div className="cart-header">
          <h3>Your Cart</h3>
          <button className="icon-btn" onClick={onClose}>✕</button>
        </div>

        {loading ? (
          <div className="cart-loading">
            <ul className="cart-list">
              {Array.from({ length: 4 }).map((_, i) => (
                <li key={i} className="cart-item skeleton-card">
                  <div className="skeleton skeleton-image" style={{ width: 90, height: 64, borderRadius: 8 }} />
                  <div style={{ flex: 1, paddingLeft: 8 }}>
                    <div className="skeleton skeleton-line short" style={{ height: 12, marginBottom: 8 }} />
                    <div className="skeleton skeleton-line" style={{ height: 10, marginBottom: 6, width: '60%' }} />
                  </div>
                  <div style={{ width: 80 }}>
                    <div className="skeleton skeleton-button" style={{ width: 72, height: 32 }} />
                  </div>
                </li>
              ))}
            </ul>
          </div>
        ) : !cart || !cart.items || cart.items.length === 0 ? (
          <div className="cart-empty">
            <p>Your cart is empty</p>
            <button onClick={onClose} className="btn-secondary">Continue Shopping</button>
          </div>
        ) : (
          <>
            <ul className="cart-list">
              {cart.items.map((it) => (
                <li key={it.id} className="cart-item">
                  <img src={`https://picsum.photos/seed/cart-${it.item_id}/120/90`} alt={it.item?.name} className="cart-thumb" />
                  <div className="cart-item-meta">
                    <strong>{it.item?.name}</strong>
                    <div className="muted">Qty: {it.quantity}</div>
                    <div className="muted">Price: ${it.item?.price?.toFixed(2)}</div>
                  </div>
                  <div className="cart-item-actions">
                    <button disabled={busy} onClick={() => removeItem(it.item_id)} className="btn-danger">Remove</button>
                  </div>
                </li>
              ))}
            </ul>

            <div className="cart-footer">
              <div className="cart-total">Total: ${total.toFixed(2)}</div>
              <div className="cart-buttons">
                <button onClick={onClose} className="btn-secondary">Close</button>
                <button
                  disabled={busy}
                  onClick={async () => {
                    setBusy(true);
                    try {
                      const res = await api.createOrder({ cart_id: cart.id });
                      toast.success(`Order successful! Order ID: ${res.data.order.id}`);
                      if (typeof onCheckout === 'function') onCheckout(res.data.order);
                    } catch (err) {
                      toast.error(err.response?.data?.error || 'Failed to create order');
                    } finally {
                      setBusy(false);
                    }
                  }}
                  className="btn-primary"
                >
                  Checkout
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default Cart;
