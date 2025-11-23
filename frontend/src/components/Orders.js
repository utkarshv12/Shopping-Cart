import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { toast } from 'react-toastify';
import './Cart.css';

function Orders({ onClose }) {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const res = await api.getUserOrders();
      setOrders(res.data.orders || []);
    } catch (err) {
      setOrders([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, []);

  const deleteOrder = async (id) => {
    if (!window.confirm('Delete this order?')) return;
    setBusy(true);
    try {
      await api.deleteOrder(id);
      toast.success('Order deleted');
      await loadOrders();
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to delete order');
    } finally {
      setBusy(false);
    }
  };

  const clearHistory = async () => {
    if (!window.confirm('Clear all order history? This cannot be undone.')) return;
    setBusy(true);
    try {
      await api.clearOrders();
      toast.success('Order history cleared');
      await loadOrders();
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to clear order history');
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <div className="cart-modal">Loading orders...</div>;

  if (!orders || orders.length === 0) {
    return (
      <div className="cart-modal">
        <h2>Your orders</h2>
        <p>You have no orders yet.</p>
        <div className="cart-footer">
          <button onClick={onClose} className="btn-secondary">Close</button>
        </div>
      </div>
    );
  }

  return (
    <div className="cart-modal">
      <div style={{display:'flex',justifyContent:'space-between',alignItems:'center'}}>
        <h2>Your Orders</h2>
        <div>
          <button className="btn-danger" onClick={clearHistory} disabled={busy}>Clear History</button>
        </div>
      </div>
      <ul className="cart-list">
        {orders.map((order) => (
          <li key={order.id} className="cart-item">
            <div>
              <strong>Order #{order.id}</strong>
              <div>Total: ${order.total.toFixed(2)}</div>
              <div>Date: {new Date(order.created_at).toLocaleString()}</div>
              <div>Items:</div>
              <ul>
                {order.items && order.items.map((it) => (
                  <li key={it.id}>
                    {it.item?.name || 'Item'} x{it.quantity} @ ${it.price.toFixed(2)}
                  </li>
                ))}
              </ul>
            </div>
            <div style={{display:'flex',alignItems:'center',gap:'8px'}}>
              <button className="btn-danger" disabled={busy} onClick={() => deleteOrder(order.id)}>Delete</button>
            </div>
          </li>
        ))}
      </ul>
      <div className="cart-footer">
        <button onClick={onClose} className="btn-secondary">Close</button>
      </div>
    </div>
  );
}

export default Orders;
