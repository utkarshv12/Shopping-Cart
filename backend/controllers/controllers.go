package controllers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"shopping-cart/config"
	"shopping-cart/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User Controllers

func CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

func GetUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func LoginUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Invalidate previous sessions for this user (single active session requirement)
	config.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})

	// Create a new session
	token := uuid.New().String()
	now := time.Now()
	// You can set an expiry if desired; leaving nil means no expiry
	session := models.Session{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: now,
	}
	if err := config.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": user.ID,
	})
}

// LogoutUser deletes the session associated with the provided token
func LogoutUser(c *gin.Context) {
	// Read Authorization header
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header"})
		return
	}
	token := parts[1]

	// Delete the session with this token
	if err := config.DB.Where("token = ?", token).Delete(&models.Session{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Item Controllers

func CreateItem(c *gin.Context) {
	var input struct {
		Name        string  `json:"name" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		Description string  `json:"description"`
		ImageData   string  `json:"image_data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.Item{
		Name:        input.Name,
		Price:       input.Price,
		Description: input.Description,
	}

	// If image data was provided (base64), decode and save it
	if strings.TrimSpace(input.ImageData) != "" {
		// ensure static/images directory exists
		imgDir := filepath.Join("static", "images")
		if err := os.MkdirAll(imgDir, 0755); err == nil {
			// image data may be like: data:image/png;base64,XXXXX -- strip prefix
			commaIndex := strings.Index(input.ImageData, ",")
			raw := input.ImageData
			if commaIndex != -1 {
				raw = input.ImageData[commaIndex+1:]
			}
			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err == nil {
				fname := uuid.New().String() + ".png"
				fpath := filepath.Join(imgDir, fname)
				if err := ioutil.WriteFile(fpath, decoded, 0644); err == nil {
					// set ImageURL to be served from /static/images/
					item.ImageURL = "/static/images/" + fname
				}
			}
		}
	}

	if err := config.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Item created successfully", "item": item})
}

func GetItems(c *gin.Context) {
	var items []models.Item
	config.DB.Find(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Cart Controllers

func AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var input struct {
		ItemID uint `json:"item_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if item exists
	var item models.Item
	if err := config.DB.First(&item, input.ItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Get or create cart for user in a transaction to avoid races
	var cart models.Cart
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start DB transaction"})
		return
	}
	if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		// not found -> create
		cart = models.Cart{UserID: userID.(uint)}
		if err := tx.Create(&cart).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
			return
		}
	}

	// Check if item already in cart
	var existingCartItem models.CartItem
	if err := tx.Where("cart_id = ? AND item_id = ?", cart.ID, input.ItemID).First(&existingCartItem).Error; err == nil {
		// Update quantity
		existingCartItem.Quantity++
		if err := tx.Save(&existingCartItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Item quantity updated in cart", "cart_item": existingCartItem})
		return
	}

	// Add item to cart
	cartItem := models.CartItem{
		CartID:   cart.ID,
		ItemID:   input.ItemID,
		Quantity: 1,
	}

	if err := tx.Create(&cartItem).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit cart transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Item added to cart", "cart_item": cartItem})
}

func GetCarts(c *gin.Context) {
	var carts []models.Cart
	config.DB.Preload("Items.Item").Preload("User").Find(&carts)
	c.JSON(http.StatusOK, gin.H{"carts": carts})
}

func GetUserCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var cart models.Cart
	if err := config.DB.Preload("Items.Item").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found", "cart": models.Cart{Items: []models.CartItem{}}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart": cart})
}

// GetCartByID returns a cart by its ID. Non-admin users may only fetch their own cart.
func GetCartByID(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	// parse id
	var cartID uint64
	var err error
	if cartID, err = strconv.ParseUint(idParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var cart models.Cart
	if err := config.DB.Preload("Items.Item").Preload("User").First(&cart, uint(cartID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Enforce ownership: only the owner can fetch (no admin concept currently)
	if cart.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cart does not belong to the authenticated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cart": cart})
}

// RemoveFromCart removes an item from the authenticated user's cart by item_id
func RemoveFromCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	itemIDParam := c.Param("item_id")
	if itemIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}

	// parse item id
	var itemID uint64
	var err error
	if itemID, err = strconv.ParseUint(itemIDParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	// Find user's cart
	var cart models.Cart
	if err := config.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Delete the cart item for this cart and item id
	if err := config.DB.Where("cart_id = ? AND item_id = ?", cart.ID, uint(itemID)).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// Order Controllers

func CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	// Accept optional cart_id in request body. If not provided, use authenticated user's cart.
	var input struct {
		CartID uint `json:"cart_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && err != http.ErrBodyNotAllowed {
		// if body present but invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart models.Cart
	var err error
	if input.CartID != 0 {
		// find cart by id and ensure it belongs to user
		err = config.DB.Preload("Items.Item").First(&cart, input.CartID).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}
		if cart.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cart does not belong to the authenticated user"})
			return
		}
	} else {
		// Get user's cart
		err = config.DB.Preload("Items.Item").Where("user_id = ?", userID).First(&cart).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}
	}

	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	// Use transaction to create order and order items, then clear cart atomically
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start DB transaction"})
		return
	}

	// Calculate total
	var total float64
	for _, cartItem := range cart.Items {
		total += cartItem.Item.Price * float64(cartItem.Quantity)
	}

	order := models.Order{
		UserID: userID.(uint),
		Total:  total,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	for _, cartItem := range cart.Items {
		orderItem := models.OrderItem{
			OrderID:  order.ID,
			ItemID:   cartItem.ItemID,
			Quantity: cartItem.Quantity,
			Price:    cartItem.Item.Price,
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order items"})
			return
		}
	}

	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Load order with items
	config.DB.Preload("Items.Item").First(&order, order.ID)

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order": order})
}

func GetOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("Items.Item").Preload("User").Find(&orders)
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var orders []models.Order
	config.DB.Preload("Items.Item").Where("user_id = ?", userID).Find(&orders)
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// DeleteOrder deletes a single order by id if it belongs to the authenticated user
func DeleteOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Items").First(&order, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Order does not belong to the authenticated user"})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start DB transaction"})
		return
	}

	if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order items"})
		return
	}
	if err := tx.Delete(&models.Order{}, order.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}

// ClearUserOrders deletes all orders (and their items) belonging to the authenticated user
func ClearUserOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var orders []models.Order
	if err := config.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query orders"})
		return
	}
	if len(orders) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No orders to delete"})
		return
	}

	ids := make([]uint, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.ID)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start DB transaction"})
		return
	}

	if err := tx.Where("order_id IN ?", ids).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order items"})
		return
	}

	if err := tx.Where("user_id = ?", userID).Delete(&models.Order{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete orders"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All orders deleted"})
}

// DeleteItem removes an item by id. Protected route.
func DeleteItem(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Check exists
	var item models.Item
	if err := config.DB.First(&item, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if err := config.DB.Delete(&models.Item{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}
