package api

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

// Admin Dashboard Page Handler
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Validate admin user (for now, any authenticated user can access admin dashboard)
	// In production, you'd want to check for admin role
	userID, validUser := ValidateUser(w, r)
	if !validUser {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	query := "SELECT COUNT(*) FROM users WHERE id = ? and user_type = 1;"
	var count int
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Database error")
		return
	}
	db.QueryRow(query, userID).Scan(&count)
	if count <= 0 {
		HandleError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}
	tmpl, err := template.ParseFiles(FrontendAdminDashboardHTML)
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load admin dashboard")
		return
	}

	data := map[string]interface{}{
		"UserID": userID,
	}

	if err := tmpl.Execute(w, data); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render admin dashboard")
		return
	}
}

// Admin Stats API
func AdminStats(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var stats struct {
		TotalStores   int     `json:"total_stores"`
		TotalUsers    int     `json:"total_users"`
		TotalProducts int     `json:"total_products"`
		TotalOrders   int     `json:"total_orders"`
		PendingOrders int     `json:"pending_orders"`
		TotalRevenue  float64 `json:"total_revenue"`
	}

	// Count stores
	db.QueryRow("SELECT COUNT(*) FROM stores").Scan(&stats.TotalStores)

	// Count users
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Count products
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&stats.TotalProducts)

	// Count orders
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)

	// Count pending orders
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE status != 'completed'").Scan(&stats.PendingOrders)

	// Get total revenue
	db.QueryRow("SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed'").Scan(&stats.TotalRevenue)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Admin Stores API
func AdminStores(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT s.id, s.name, s.template, u.first_name || ' ' || u.last_name, COUNT(p.id) as product_count
		FROM stores s
		LEFT JOIN users u ON s.owner_id = u.id
		LEFT JOIN products p ON s.id = p.store_id
		GROUP BY s.id
		ORDER BY s.id DESC
	`)
	if err != nil {
		log.Println("Error querying stores:", err)
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stores []map[string]interface{}
	for rows.Next() {
		var id int
		var name, template, owner string
		var productCount int

		if err := rows.Scan(&id, &name, &template, &owner, &productCount); err != nil {
			continue
		}

		// Fetch store and set cookie
		store, err := GetStoreByID(id)
		if err == nil && store.ID != 0 {
			SetStoreCookie(w, int(store.OwnerID), store.Name)
		}

		stores = append(stores, map[string]interface{}{
			"id":            id,
			"name":          name,
			"template":      template,
			"owner":         owner,
			"product_count": productCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if stores == nil {
		stores = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(stores)
}

// Admin Users API
func AdminUsers(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT u.id,  u.first_name || ' ' || u.last_name, u.email, u.created_at, COUNT(s.id) as store_count
		FROM users u
		LEFT JOIN stores s ON u.id = s.owner_id
		GROUP BY u.id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		log.Println("Error querying users:", err)
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email, createdAt string
		var storeCount int

		if err := rows.Scan(&id, &name, &email, &createdAt, &storeCount); err != nil {
			continue
		}

		users = append(users, map[string]interface{}{
			"id":          id,
			"name":        name,
			"email":       email,
			"created_at":  createdAt,
			"store_count": storeCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if users == nil {
		users = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(users)
}

// Admin Orders API
func AdminOrders(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT o.id, o.created_at, o.total_amount, o.status, s.name, COALESCE(u.first_name || ' ' || u.last_name, u.email, 'Guest')
		FROM orders o
		LEFT JOIN stores s ON o.store_id = s.id
		LEFT JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT 100
	`)
	if err != nil {
		log.Println("Error querying orders:", err)
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var id int
		var orderDate, status, customerName string
		var totalAmount float64
		var storeNamePtr *string

		if err := rows.Scan(&id, &orderDate, &totalAmount, &status, &storeNamePtr, &customerName); err != nil {
			log.Println("Error scanning order row:", err)
			continue
		}

		storeNameVal := "N/A"
		if storeNamePtr != nil {
			storeNameVal = *storeNamePtr
		}

		orders = append(orders, map[string]interface{}{
			"id":            id,
			"created_at":    orderDate,
			"total":         totalAmount,
			"status":        status,
			"store_name":    storeNameVal,
			"customer_name": customerName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if orders == nil {
		orders = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(orders)
}

// Admin Products API
func AdminProducts(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT p.id, p.name, p.price, p.quantity, s.name
		FROM products p
		LEFT JOIN stores s ON p.store_id = s.id
		ORDER BY p.id DESC
		LIMIT 100
	`)
	if err != nil {
		log.Println("Error querying products:", err)
		http.Error(w, `[]`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var name, storeName string
		var price float64
		var quantity int

		if err := rows.Scan(&id, &name, &price, &quantity, &storeName); err != nil {
			continue
		}

		products = append(products, map[string]interface{}{
			"id":         id,
			"name":       name,
			"price":      price,
			"quantity":   quantity,
			"store_name": storeName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if products == nil {
		products = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(products)
}
