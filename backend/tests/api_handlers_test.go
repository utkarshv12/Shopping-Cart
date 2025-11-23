package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"shopping-cart/config"
	"shopping-cart/models"
	"shopping-cart/routes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var router *gin.Engine

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)

	// Initialize DB for tests. Prefer MySQL if TEST_MYSQL_DSN or MYSQL_DSN is set.
	var db *gorm.DB
	var err error
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	// Require a MySQL DSN for tests in this project (no sqlite fallback)
	if dsn == "" {
		Fail("TEST_MYSQL_DSN or MYSQL_DSN must be set to run tests (this project uses MySQL for tests).")
		return
	}

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	Expect(err).To(BeNil())
	config.DB = db

	// Migrate
	err = db.AutoMigrate(&models.User{}, &models.Item{}, &models.Cart{}, &models.CartItem{}, &models.Order{}, &models.OrderItem{}, &models.Session{})
	Expect(err).To(BeNil())

	router = gin.Default()
	routes.SetupRoutes(router)
})

var _ = Describe("API End-to-end", func() {
	It("should allow signup, login, create item, add to cart and checkout", func() {
		// Signup
		signupBody := map[string]string{"username": "testuser", "password": "password123"}
		w := performRequest(router, "POST", "/users", signupBody, "")
		Expect(w.Code).To(Equal(http.StatusCreated))

		// Login
		loginBody := map[string]string{"username": "testuser", "password": "password123"}
		w = performRequest(router, "POST", "/users/login", loginBody, "")
		Expect(w.Code).To(Equal(http.StatusOK))

		var loginResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &loginResp)
		token, ok := loginResp["token"].(string)
		Expect(ok).To(BeTrue())

		// Create Item
		itemBody := map[string]interface{}{"name": "Widget", "price": 9.99, "description": "A test widget"}
		w = performRequest(router, "POST", "/items", itemBody, "")
		Expect(w.Code).To(Equal(http.StatusCreated))

		// List items to get ID
		w = performRequest(router, "GET", "/items", nil, "")
		Expect(w.Code).To(Equal(http.StatusOK))
		var itemsResp map[string][]map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &itemsResp)
		Expect(len(itemsResp["items"])).To(BeNumerically(">", 0))
		itemID := uint(itemsResp["items"][0]["id"].(float64))

		// Add to cart
		cartBody := map[string]uint{"item_id": itemID}
		w = performRequest(router, "POST", "/carts", cartBody, token)
		Expect(w.Code == http.StatusCreated || w.Code == http.StatusOK).To(BeTrue())

		// Get user cart
		w = performRequest(router, "GET", "/carts/user", nil, token)
		Expect(w.Code).To(Equal(http.StatusOK))
		var cartResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &cartResp)
		cart := cartResp["cart"].(map[string]interface{})
		Expect(cart).NotTo(BeNil())

		// Checkout -> create order
		w = performRequest(router, "POST", "/orders", nil, token)
		Expect(w.Code).To(Equal(http.StatusCreated))

		// Get user orders
		w = performRequest(router, "GET", "/orders/user", nil, token)
		Expect(w.Code).To(Equal(http.StatusOK))
		var ordersResp map[string][]map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &ordersResp)
		Expect(len(ordersResp["orders"])).To(BeNumerically(">", 0))
	})
})

func performRequest(router *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
