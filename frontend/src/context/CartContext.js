import React, { createContext, useContext, useEffect, useState } from 'react';
import api from '../services/api';

const CartContext = createContext(null);

export function useCart() {
  return useContext(CartContext);
}

export function CartProvider({ children }) {
  const [cart, setCart] = useState(null);
  const [loading, setLoading] = useState(false);

  const refresh = async () => {
    setLoading(true);
    try {
      const res = await api.getUserCart();
      setCart(res.data.cart || null);
    } catch (err) {
      setCart(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // initial fetch
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const value = {
    cart,
    loading,
    refresh,
    count: cart && cart.items ? cart.items.reduce((s, it) => s + (it.quantity || 0), 0) : 0,
  };

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export default CartContext;
