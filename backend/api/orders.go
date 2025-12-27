package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get shipping info from form
	fullName := strings.TrimSpace(r.FormValue("full_name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	address := strings.TrimSpace(r.FormValue("address"))
	city := strings.TrimSpace(r.FormValue("city"))
	state := strings.TrimSpace(r.FormValue("state"))
	zipCode := strings.TrimSpace(r.FormValue("zip_code"))
	country := strings.TrimSpace(r.FormValue("country"))

	if fullName == "" || address == "" || city == "" || country == "" {
		http.Error(w, `{"error": "All shipping fields are required"}`, http.StatusBadRequest)
		return
	}

	shippingInfo := strings.Join([]string{fullName, phone, address, city, state, zipCode, country}, "|")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get cart items and calculate total
	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.quantity, p.price, p.store_id
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ?
	`, userID)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch cart"}`, http.StatusInternalServerError)
		return
	}

	type CartData struct {
		ID        int
		ProductID int
		Quantity  int
		Price     float64
		StoreID   int
	}

	var cartItems []CartData
	var storeID int
	totalAmount := 0.0

	for rows.Next() {
		var item CartData
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.StoreID); err != nil {
			rows.Close()
			http.Error(w, `{"error": "Failed to process cart"}`, http.StatusInternalServerError)
			return
		}
		if storeID == 0 {
			storeID = item.StoreID
		}
		totalAmount += item.Price * float64(item.Quantity)
		cartItems = append(cartItems, item)
	}
	rows.Close()

	if len(cartItems) == 0 {
		http.Error(w, `{"error": "Cart is empty"}`, http.StatusBadRequest)
		return
	}

	// Delete any older pending orders (unpaid orders)
	_, err = db.Exec("DELETE FROM orders WHERE user_id = ? AND status = 'pending'", userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to clean up old orders"}`, http.StatusInternalServerError)
		return
	}
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	// Create order
	result, err := tx.Exec(`
		INSERT INTO orders (user_id, store_id, total_amount, status, shipping_info, created_at)
		VALUES (?, ?, ?, 'pending', ?, datetime('now'))
	`, userID, storeID, totalAmount, shippingInfo)

	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to create order"}`, http.StatusInternalServerError)
		return
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to create order"}`, http.StatusInternalServerError)
		return
	}

	// Create order items
	for _, item := range cartItems {
		_, err = tx.Exec(`
			INSERT INTO order_products (order_id, product_id, quantity, price)
			VALUES (?, ?, ?, ?)
		`, orderID, item.ProductID, item.Quantity, item.Price)

		if err != nil {
			tx.Rollback()
			http.Error(w, `{"error": "Failed to create order items"}`, http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, `{"error": "Failed to complete order"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"order_id": orderID,
		"message":  "Order created successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, user_id, store_id, total_amount, status, shipping_info, created_at
		FROM orders
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch orders"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []map[string]interface{}

	for rows.Next() {
		var id, userID, storeID int
		var totalAmount float64
		var status, shippingInfo, createdAt string

		err := rows.Scan(&id, &userID, &storeID, &totalAmount, &status, &shippingInfo, &createdAt)
		if err != nil {
			continue
		}

		orders = append(orders, map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"store_id":      storeID,
			"total_amount":  totalAmount,
			"status":        status,
			"shipping_info": shippingInfo,
			"created_at":    createdAt,
		})
	}

	json.NewEncoder(w).Encode(orders)
}

// GetPendingOrder retrieves the most recent pending order for payment
func GetPendingOrder(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get the most recent pending order
	var orderID, storeID int
	var totalAmount float64
	var shippingInfo, createdAt string

	err = db.QueryRow(`
		SELECT id, store_id, total_amount, shipping_info, created_at
		FROM orders
		WHERE user_id = ? AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&orderID, &storeID, &totalAmount, &shippingInfo, &createdAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "No pending order found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to fetch order"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"order": map[string]interface{}{
			"order_id":      orderID,
			"store_id":      storeID,
			"total_amount":  totalAmount,
			"shipping_info": shippingInfo,
			"created_at":    createdAt,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// ProcessPayment simulates payment processing (fake payment system)
func ProcessPayment(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get form data
	orderIDStr := strings.TrimSpace(r.FormValue("order_id"))
	cardHolder := strings.TrimSpace(r.FormValue("card_holder"))
	cardNumber := strings.TrimSpace(r.FormValue("card_number"))
	expiry := strings.TrimSpace(r.FormValue("expiry"))
	cvv := strings.TrimSpace(r.FormValue("cvv"))

	// Validate inputs
	if orderIDStr == "" || cardHolder == "" || cardNumber == "" || expiry == "" || cvv == "" {
		http.Error(w, `{"error": "All payment fields are required"}`, http.StatusBadRequest)
		return
	}

	// Convert order ID
	orderID := 0
	if _, err := fmt.Sscanf(orderIDStr, "%d", &orderID); err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	// Basic card validation
	cardNumberClean := strings.ReplaceAll(cardNumber, " ", "")
	if len(cardNumberClean) < 13 || len(cardNumberClean) > 19 {
		http.Error(w, `{"error": "Invalid card number"}`, http.StatusBadRequest)
		return
	}

	// Validate CVV
	if len(cvv) < 3 || len(cvv) > 4 {
		http.Error(w, `{"error": "Invalid CVV"}`, http.StatusBadRequest)
		return
	}

	// Validate expiry format (MM/YY)
	if len(expiry) != 5 || expiry[2] != '/' {
		http.Error(w, `{"error": "Invalid expiry date format"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Verify order belongs to user and is pending
	var orderUserID int
	var status string
	err = db.QueryRow(`
		SELECT user_id, status
		FROM orders
		WHERE id = ?
	`, orderID).Scan(&orderUserID, &status)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to verify order"}`, http.StatusInternalServerError)
		return
	}

	if orderUserID != userID {
		http.Error(w, `{"error": "Unauthorized to pay for this order"}`, http.StatusForbidden)
		return
	}

	if status != "pending" {
		http.Error(w, `{"error": "Order has already been processed"}`, http.StatusBadRequest)
		return
	}

	// Simulate payment processing (fake payment - always succeeds)
	// In a real system, this would integrate with a payment gateway like Stripe, PayPal, etc.

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	// Store payment information (in real system, never store full card details!)
	// For demo purposes, we'll just store masked card info
	maskedCard := "****" + cardNumberClean[len(cardNumberClean)-4:]
	paymentMethodName := "Credit Card (" + maskedCard + ")"

	// Insert into payment_methods table
	paymentResult, err := tx.Exec(`
		INSERT INTO payment_methods (user_id, method_name, account_details, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`, userID, paymentMethodName, maskedCard)

	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to store payment method"}`, http.StatusInternalServerError)
		return
	}

	paymentMethodID, err := paymentResult.LastInsertId()
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to retrieve payment method ID"}`, http.StatusInternalServerError)
		return
	}

	// Get order total amount
	var totalAmount float64
	err = tx.QueryRow(`SELECT total_amount FROM orders WHERE id = ?`, orderID).Scan(&totalAmount)
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to retrieve order amount"}`, http.StatusInternalServerError)
		return
	}

	// Insert into transactions table
	_, err = tx.Exec(`
		INSERT INTO transactions (order_id, payment_method_id, amount, transaction_date, status)
		VALUES (?, ?, ?, datetime('now'), 'completed')
	`, orderID, paymentMethodID, totalAmount)

	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to create transaction record"}`, http.StatusInternalServerError)
		return
	}

	// Store payment info in order and update status to 'paid'
	paymentInfo := strings.Join([]string{cardHolder, maskedCard, expiry}, "|")
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = 'paid', payment_info = ?, updated_at = datetime('now')
		WHERE id = ?
	`, paymentInfo, orderID)

	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to update order status"}`, http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, `{"error": "Failed to complete payment"}`, http.StatusInternalServerError)
		return
	}

	// Clear cart after successful payment
	_, err = db.Exec("DELETE FROM cart WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to clear cart"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "Payment processed successfully",
		"order_id": orderID,
	}

	json.NewEncoder(w).Encode(response)
}

// GetStoreOrders retrieves all orders for products from a specific store
func GetStoreOrders(w http.ResponseWriter, r *http.Request) {
	user, err := GetCookie(r, "session_token")
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get user ID from session
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_token = ?", user).Scan(&userID)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get store ID for this user
	storeIDStr := r.URL.Query().Get("store_id")
	storeID := 0
	if _, err := fmt.Sscanf(storeIDStr, "%d", &storeID); err != nil {
		http.Error(w, `{"error": "Invalid store ID"}`, http.StatusBadRequest)
		return
	}

	// Verify user owns this store
	var storeOwnerID int
	err = db.QueryRow("SELECT owner_id FROM stores WHERE id = ?", storeID).Scan(&storeOwnerID)
	if err != nil || storeOwnerID != userID {
		http.Error(w, `{"error": "Unauthorized to view these orders"}`, http.StatusUnauthorized)
		return
	}

	// Get all orders for this store
	query := `
		SELECT 
			o.id, 
			o.user_id, 
			o.total_amount, 
			o.status, 
			o.shipping_info, 
			o.payment_info, 
			o.created_at, 
			o.updated_at,
			u.first_name,
			COUNT(op.id) as product_count
		FROM orders o
		LEFT JOIN order_products op ON o.id = op.order_id
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.store_id = ?
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`

	rows, err := db.Query(query, storeID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch orders"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrderData struct {
		ID           int    `json:"id"`
		CustomerName string `json:"customer_name"`
		ProductCount int    `json:"product_count"`
		TotalAmount  string `json:"total_amount"`
		Status       string `json:"status"`
		CreatedAt    string `json:"created_at"`
	}

	var orders []OrderData
	for rows.Next() {
		var order OrderData
		var userID int
		var firstName sql.NullString
		var shippingInfo sql.NullString
		var paymentInfo sql.NullString
		var updatedAt sql.NullString

		err := rows.Scan(
			&order.ID,
			&userID,
			&order.TotalAmount,
			&order.Status,
			&shippingInfo,
			&paymentInfo,
			&order.CreatedAt,
			&updatedAt,
			&firstName,
			&order.ProductCount,
		)
		if err != nil {
			continue
		}

		if firstName.Valid {
			order.CustomerName = firstName.String
		} else {
			order.CustomerName = "Guest"
		}

		orders = append(orders, order)
	}

	response := map[string]interface{}{
		"orders": orders,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateOrderStatus updates the status of an order
func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	user, err := GetCookie(r, "session_token")
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		OrderID int    `json:"order_id"`
		Status  string `json:"status"`
		StoreID int    `json:"store_id"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get user ID from session
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_token = ?", user).Scan(&userID)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Verify user owns this store
	var storeOwnerID int
	err = db.QueryRow("SELECT owner_id FROM stores WHERE id = ?", req.StoreID).Scan(&storeOwnerID)
	if err != nil || storeOwnerID != userID {
		http.Error(w, `{"error": "Unauthorized to update this order"}`, http.StatusUnauthorized)
		return
	}

	// Verify the order belongs to this store
	var orderExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM orders
		WHERE id = ? AND store_id = ?
	`, req.OrderID, req.StoreID).Scan(&orderExists)
	if err != nil || orderExists == 0 {
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	}

	// Update order status
	_, err = db.Exec(
		"UPDATE orders SET status = ?, updated_at = datetime('now') WHERE id = ?",
		req.Status,
		req.OrderID,
	)
	if err != nil {
		http.Error(w, `{"error": "Failed to update order"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Order status updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetOrderProducts retrieves all products in a specific order
func GetOrderProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get store ID and order ID from URL
	storeIDStr := r.URL.Query().Get("store_id")
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/orders/"), "/")
	if len(pathSegments) < 1 {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	orderIDStr := pathSegments[0]
	storeID := 0
	orderID := 0
	if _, err := fmt.Sscanf(storeIDStr, "%d", &storeID); err != nil {
		http.Error(w, `{"error": "Invalid store ID"}`, http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(orderIDStr, "%d", &orderID); err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	// Try to validate as store owner first
	userID, isStoreOwner := ValidateUser(w, r)
	isCustomer := false
	customerID := 0

	if !isStoreOwner {
		// If not store owner, try to validate as customer
		customerID, isCustomer = ValidateCustomer(w, r)
		if !isCustomer {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
	}

	// If store owner, verify they own this store
	if isStoreOwner {
		var storeOwnerID int
		err = db.QueryRow("SELECT owner_id FROM stores WHERE id = ?", storeID).Scan(&storeOwnerID)
		if err != nil || storeOwnerID != userID {
			http.Error(w, `{"error": "Unauthorized to view these products"}`, http.StatusUnauthorized)
			return
		}

		// Verify the order belongs to this store
		var orderExists int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM orders
			WHERE id = ? AND store_id = ?
		`, orderID, storeID).Scan(&orderExists)
		if err != nil || orderExists == 0 {
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
			return
		}
	} else if isCustomer {
		// If customer, verify the order belongs to them and this store
		var orderUserID int
		err = db.QueryRow(`
			SELECT user_id FROM orders
			WHERE id = ? AND store_id = ?
		`, orderID, storeID).Scan(&orderUserID)
		if err != nil || orderUserID != customerID {
			http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
			return
		}
	}

	// Get all products in this order
	query := `
		SELECT op.product_id, p.name, p.image, op.quantity, op.price
		FROM order_products op
		INNER JOIN products p ON op.product_id = p.id
		INNER JOIN orders o ON op.order_id = o.id
		INNER JOIN stores s ON p.store_id = s.id
		WHERE o.id = ? AND s.id = ?
		ORDER BY op.id
	`

	rows, err := db.Query(query, orderID, storeID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch products"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrderProduct struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Image    string  `json:"image"`
		Quantity int     `json:"quantity"`
		Price    float64 `json:"price"`
	}

	var products []OrderProduct
	for rows.Next() {
		var product OrderProduct
		err := rows.Scan(&product.ID, &product.Name, &product.Image, &product.Quantity, &product.Price)
		if err != nil {
			continue
		}
		products = append(products, product)
	}

	response := map[string]interface{}{
		"products": products,
	}

	json.NewEncoder(w).Encode(response)
}

// GetCustomerOrders fetches all orders for the logged-in customer for a specific store
func GetCustomerOrders(w http.ResponseWriter, r *http.Request) {
	storeIDStr := r.URL.Query().Get("store_id")
	if storeIDStr == "" {
		http.Error(w, `{"error": "store_id is required"}`, http.StatusBadRequest)
		return
	}
	userID, validUser := ValidateStoreCustomer(w, r, storeIDStr)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get orders for this customer on this store
	query := `
		SELECT 
			o.id,
			o.created_at,
			o.status,
			SUM(op.quantity * op.price) as total_amount,
			COUNT(op.id) as product_count
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		JOIN products p ON op.product_id = p.id
		WHERE o.user_id = ? AND p.store_id = ?
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`

	rows, err := db.Query(query, userID, storeIDStr)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch orders"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrderInfo struct {
		ID           int     `json:"id"`
		CreatedAt    string  `json:"created_at"`
		Status       string  `json:"status"`
		TotalAmount  float64 `json:"total_amount"`
		ProductCount int     `json:"product_count"`
	}

	var orders []OrderInfo
	for rows.Next() {
		var order OrderInfo
		err := rows.Scan(&order.ID, &order.CreatedAt, &order.Status, &order.TotalAmount, &order.ProductCount)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	response := map[string]interface{}{
		"orders": orders,
	}

	json.NewEncoder(w).Encode(response)
}
