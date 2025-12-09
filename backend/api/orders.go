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
	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	address := strings.TrimSpace(r.FormValue("address"))
	city := strings.TrimSpace(r.FormValue("city"))
	state := strings.TrimSpace(r.FormValue("state"))
	zipCode := strings.TrimSpace(r.FormValue("zip_code"))
	country := strings.TrimSpace(r.FormValue("country"))

	if fullName == "" || email == "" || address == "" || city == "" || country == "" {
		http.Error(w, `{"error": "All shipping fields are required"}`, http.StatusBadRequest)
		return
	}

	shippingInfo := strings.Join([]string{fullName, email, phone, address, city, state, zipCode, country}, "|")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get cart items and calculate total
	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.quantity, p.price
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
	}

	var cartItems []CartData
	totalAmount := 0.0

	for rows.Next() {
		var item CartData
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			rows.Close()
			http.Error(w, `{"error": "Failed to process cart"}`, http.StatusInternalServerError)
			return
		}
		totalAmount += item.Price * float64(item.Quantity)
		cartItems = append(cartItems, item)
	}
	rows.Close()

	if len(cartItems) == 0 {
		http.Error(w, `{"error": "Cart is empty"}`, http.StatusBadRequest)
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
		INSERT INTO orders (user_id, total_amount, status, shipping_info, created_at)
		VALUES (?, ?, 'pending', ?, datetime('now'))
	`, userID, totalAmount, shippingInfo)

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

	// Clear cart
	_, err = tx.Exec("DELETE FROM cart WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Failed to clear cart"}`, http.StatusInternalServerError)
		return
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
	userID, validUser := ValidateUser(w, r)
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
		SELECT id, user_id, total_amount, status, shipping_info, created_at
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
		var id, userID int
		var totalAmount float64
		var status, shippingInfo, createdAt string

		err := rows.Scan(&id, &userID, &totalAmount, &status, &shippingInfo, &createdAt)
		if err != nil {
			continue
		}

		orders = append(orders, map[string]interface{}{
			"id":            id,
			"user_id":       userID,
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
	var orderID int
	var totalAmount float64
	var shippingInfo, createdAt string

	err = db.QueryRow(`
		SELECT id, total_amount, shipping_info, created_at
		FROM orders
		WHERE user_id = ? AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&orderID, &totalAmount, &shippingInfo, &createdAt)

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

	response := map[string]interface{}{
		"success":  true,
		"message":  "Payment processed successfully",
		"order_id": orderID,
	}

	json.NewEncoder(w).Encode(response)
}
