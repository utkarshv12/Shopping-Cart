package routes

import (
	"shopping-cart/controllers"
	"shopping-cart/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// User routes
	r.POST("/users", controllers.CreateUser)
	r.GET("/users", controllers.GetUsers)
	r.POST("/users/login", controllers.LoginUser)

	// Item routes
	r.POST("/items", controllers.CreateItem)
	r.GET("/items", controllers.GetItems)

	// Protected routes (require authentication)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		// User logout
		authorized.POST("/users/logout", controllers.LogoutUser)
		// Cart routes
		authorized.POST("/carts", controllers.AddToCart)
		authorized.DELETE("/carts/items/:item_id", controllers.RemoveFromCart)
		// Item delete (protected)
		authorized.DELETE("/items/:id", controllers.DeleteItem)
		authorized.GET("/carts/:id", controllers.GetCartByID)
		authorized.GET("/carts", controllers.GetCarts)
		authorized.GET("/carts/user", controllers.GetUserCart)

		// Order routes
		authorized.POST("/orders", controllers.CreateOrder)
		// Delete single order
		authorized.DELETE("/orders/:id", controllers.DeleteOrder)
		// Clear all user orders
		authorized.DELETE("/orders/user", controllers.ClearUserOrders)
		authorized.GET("/orders", controllers.GetOrders)
		authorized.GET("/orders/user", controllers.GetUserOrders)
	}
}
