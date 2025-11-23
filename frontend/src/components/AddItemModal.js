import React, { useState } from 'react';
import api from '../services/api';
import { toast } from 'react-toastify';
import './Cart.css';

function AddItemModal({ onClose, onCreated }) {
  const [name, setName] = useState('');
  const [price, setPrice] = useState('');
  const [description, setDescription] = useState('');
  const [busy, setBusy] = useState(false);
  const [imagePreview, setImagePreview] = useState(null);
  const [imageFile, setImageFile] = useState(null);

  const toBase64 = (file) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = (error) => reject(error);
  });

  const submit = async (e) => {
    e.preventDefault();
    setBusy(true);
    try {
      const priceNum = parseFloat(price);
      if (!name || isNaN(priceNum)) {
        toast.error('Please provide valid name and price');
        setBusy(false);
        return;
      }

      let payload = { name, price: priceNum, description };
      if (imageFile) {
        const b64 = await toBase64(imageFile);
        // b64 is like data:image/png;base64,.... we send it as image_data
        payload.image_data = b64;
      }

      const res = await api.createItem(payload);
      toast.success('Item created');
      setName('');
      setPrice('');
      setDescription('');
      setImageFile(null);
      setImagePreview(null);
      if (typeof onCreated === 'function') onCreated(res.data.item);
      onClose();
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to create item');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="cart-modal">
      <h2>Add Item</h2>
      <form onSubmit={submit} className="add-item-form">
        <div className="form-group">
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="form-group">
          <label>Price</label>
          <input value={price} onChange={(e) => setPrice(e.target.value)} required />
        </div>
        <div className="form-group">
          <label>Description</label>
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        <div className="form-group">
          <label>Picture (optional)</label>
          <input type="file" accept="image/*" onChange={async (e) => {
            const f = e.target.files[0];
            if (!f) return;
            setImageFile(f);
            try {
              const b64 = await toBase64(f);
              setImagePreview(b64);
            } catch (err) {
              setImagePreview(null);
            }
          }} />
          {imagePreview && <img src={imagePreview} alt="preview" style={{ width: 180, marginTop: 10, borderRadius: 8 }} />}
        </div>
        <div className="cart-footer">
          <button type="button" onClick={onClose} className="btn-secondary">Cancel</button>
          <button type="submit" className="btn-primary" disabled={busy}>{busy ? 'Please wait...' : 'Create'}</button>
        </div>
      </form>
    </div>
  );
}

export default AddItemModal;
