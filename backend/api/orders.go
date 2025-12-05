package api

import (
	"database/sql"
	"encoding/json"
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
