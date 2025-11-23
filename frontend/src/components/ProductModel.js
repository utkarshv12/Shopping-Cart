import React from 'react';
import './ProductModal.css';

function ProductModal({ item, onClose }) {
  if (!item) return null;

  return (
    <div className="pm-overlay" onClick={onClose}>
      <div className="pm-card glass" onClick={(e) => e.stopPropagation()}>
        <div className="pm-top">
          <img src={item.image_url ? `${process.env.REACT_APP_API_BASE || 'http://localhost:8080'}${item.image_url}` : `https://picsum.photos/seed/product-${item.id}/800/480`} alt={item.name} />
        </div>
        <div className="pm-body">
          <h2>{item.name}</h2>
          <p className="pm-price">${item.price?.toFixed(2)}</p>
          <p className="pm-desc">{item.description}</p>
          <div className="pm-actions">
            <button className="btn-secondary" onClick={onClose}>Close</button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ProductModal;
