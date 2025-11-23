package models

import (
	"time"

	"gorm.io/gorm"
)

// User model
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"unique;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Item model
type Item struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Price       float64        `gorm:"not null" json:"price"`
	Description string         `json:"description"`
	ImageURL    string         `json:"image_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Cart model
type Cart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;unique" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []CartItem     `gorm:"foreignKey:CartID" json:"items,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// CartItem - junction table for Cart and Item
type CartItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CartID    uint           `gorm:"not null" json:"cart_id"`
	ItemID    uint           `gorm:"not null" json:"item_id"`
	Item      Item           `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Quantity  int            `gorm:"default:1" json:"quantity"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Order model
type Order struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Total     float64        `json:"total"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// OrderItem - stores items in an order
type OrderItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OrderID   uint           `gorm:"not null" json:"order_id"`
	ItemID    uint           `gorm:"not null" json:"item_id"`
	Item      Item           `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Quantity  int            `gorm:"default:1" json:"quantity"`
	Price     float64        `gorm:"not null" json:"price"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Session represents a user's active session / token. We keep it separate
// so a user can have session metadata managed independently and easily
// invalidate previous tokens when logging in from a new device.
type Session struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	Token     string     `gorm:"unique;not null" json:"token"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
