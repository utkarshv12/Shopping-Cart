import axios from 'axios';

const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080';

const client = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
});

// attach token if present
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

const api = {
  // Users
  signup: (username, password) => client.post('/users', { username, password }),
  login: (username, password) => client.post('/users/login', { username, password }),
  logout: () => client.post('/users/logout'),
  getUsers: () => client.get('/users'),

  // Items
  createItem: (item) => client.post('/items', item),
  getItems: () => client.get('/items'),
  deleteItem: (id) => client.delete(`/items/${id}`),

  // Cart
  addToCart: (item_id) => client.post('/carts', { item_id }),
  removeFromCart: (item_id) => client.delete(`/carts/items/${item_id}`),
  getCarts: () => client.get('/carts'),
  getUserCart: () => client.get('/carts/user'),

  // Orders
  // createOrder accepts optional payload, e.g. { cart_id: 123 }
  createOrder: (payload = {}) => client.post('/orders', payload),
  getOrders: () => client.get('/orders'),
  getUserOrders: () => client.get('/orders/user'),
  deleteOrder: (id) => client.delete(`/orders/${id}`),
  clearOrders: () => client.delete('/orders/user'),
};

export default api;
