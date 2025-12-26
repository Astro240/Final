package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetProductsByStoreID(storeID uint) ([]Product, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, name, description, price, image, quantity FROM products WHERE store_id = ?", storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Image, &product.Quantity); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func CreateProductAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// Parse multipart form to access form fields and files
	r.ParseMultipartForm(10 << 20) // 10MB max

	user, err := GetCookie(r, "session_token")
	if err != nil {
		http.Error(w, `{"error": "Couldnt verify user"}`, http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database Error"}`, http.StatusInternalServerError)
		return
	}
	query := "SELECT user_id FROM sessions WHERE session_token = ?"
	defer db.Close()
	var userID int
	err = db.QueryRow(query, user).Scan(&userID)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	itemName := r.FormValue("productName")
	if itemName == "" {
		http.Error(w, `{"error": "Product name is required"}`, http.StatusInternalServerError)
		return
	}
	itemDescription := r.FormValue("productDescription")
	if itemDescription == "" {
		http.Error(w, `{"error": "Product description is required"}`, http.StatusInternalServerError)
		return
	}
	itemPrice := r.FormValue("productPrice")
	if itemPrice == "" {
		http.Error(w, `{"error": "Product price is required"}`, http.StatusInternalServerError)
		return
	}
	if price, err := strconv.ParseFloat(itemPrice, 64); err != nil || price < 0 {
		http.Error(w, `{"error": "Product price cannot be negative"}`, http.StatusInternalServerError)
		return
	}
	itemImage := r.FormValue("productImage")
	if itemImage == "" {
		http.Error(w, `{"error": "Product image is required"}`, http.StatusInternalServerError)
		return
	}
	itemQuantity := r.FormValue("productQuantity")
	if itemQuantity == "" {
		http.Error(w, `{"error": "Product quantity is required"}`, http.StatusInternalServerError)
		return
	}
	if qty, err := strconv.Atoi(itemQuantity); err != nil || qty < 0 {
		http.Error(w, `{"error": "Product quantity cannot be negative"}`, http.StatusInternalServerError)
		return
	}
	storeID := r.FormValue("store_id")
	query = "select id from stores where owner_id = ? and id = ?"
	var storeOwnerID int
	err = db.QueryRow(query, userID, storeID).Scan(&storeOwnerID)
	if err != nil {
		http.Error(w, `{"error": "Unauthorized to add products to this store"}`, http.StatusUnauthorized)
		return
	}
	var imageName string
	valid, imageName := ValidateImage(StoreProductsPath, itemImage)
	if !valid {
		http.Error(w, `{"error": "`+imageName+`"}`, http.StatusInternalServerError)
		return
	}
	insertQuery := "INSERT INTO products (name, description, price, image, quantity, store_id) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(insertQuery, itemName, itemDescription, itemPrice, imageName, itemQuantity, storeOwnerID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create product"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"success": true}`))
}

func FavoriteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Only Post Method Allowed"}`))
		return
	}
	userid, valid := ValidateCustomer(w, r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Failed to validated customer","invalid": true}`))
		return
	}
	product_id := r.FormValue("product_id")
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database Error"}`))
		return
	}
	defer db.Close()
	query := "INSERT INTO favorites (user_id,product_id) values(?,?);"
	_, err = db.Exec(query, userid, product_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to favorite product"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func UnfavoriteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Only Post Method Allowed"}`))
		return
	}
	userid, valid := ValidateCustomer(w, r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Failed to validated customer","invalid": true}`))
		return
	}
	product_id := r.FormValue("product_id")
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database Error"}`))
		return
	}
	defer db.Close()
	query := "DELETE FROM favorites WHERE user_id = ? AND product_id = ?;"
	result, err := db.Exec(query, userid, product_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to unfavorite product"}`))
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Favorite not found"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// GetProductsAPI - Get all products for a store (JSON response)
func GetProductsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	storeIDStr := r.URL.Query().Get("store_id")
	if storeIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "store_id is required"})
		return
	}

	storeID, err := strconv.ParseUint(storeIDStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid store_id"})
		return
	}

	products, err := GetProductsByStoreID(uint(storeID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch products"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"products": products,
		"success":  true,
	})
}

// UpdateProductAPI - Update an existing product
func UpdateProductAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only POST method allowed"})
		return
	}

	// Get user from session
	user, err := GetCookie(r, "session_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer db.Close()

	var userID int
	query := "SELECT user_id FROM sessions WHERE session_token = ?"
	err = db.QueryRow(query, user).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Parse JSON request
	var productReq struct {
		ID          uint    `json:"id"`
		StoreID     uint    `json:"store_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Quantity    int     `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&productReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate inputs
	if productReq.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product ID is required"})
		return
	}

	if productReq.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product name is required"})
		return
	}

	if productReq.Price < 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product price cannot be negative"})
		return
	}

	if productReq.Quantity < 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product quantity cannot be negative"})
		return
	}

	// Verify product exists and get store_id
	var checkStoreID uint
	query = "SELECT store_id FROM products WHERE id = ?"
	err = db.QueryRow(query, productReq.ID).Scan(&checkStoreID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}

	// Verify user owns the store
	var storeOwnerID int
	query = "SELECT id FROM stores WHERE owner_id = ? AND id = ?"
	err = db.QueryRow(query, userID, checkStoreID).Scan(&storeOwnerID)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized to update this product"})
		return
	}

	// Update product
	updateQuery := "UPDATE products SET name = ?, description = ?, price = ?, quantity = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, productReq.Name, productReq.Description, productReq.Price, productReq.Quantity, productReq.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update product"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// DeleteProductAPI - Delete a product
func DeleteProductAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only POST method allowed"})
		return
	}

	// Get user from session
	user, err := GetCookie(r, "session_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer db.Close()

	var userID int
	query := "SELECT user_id FROM sessions WHERE session_token = ?"
	err = db.QueryRow(query, user).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Parse JSON request
	var deleteReq struct {
		ID      uint `json:"id"`
		StoreID uint `json:"store_id"`
	}

	err = json.NewDecoder(r.Body).Decode(&deleteReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if deleteReq.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product ID is required"})
		return
	}

	// Verify product exists and get store_id
	var checkStoreID uint
	query = "SELECT store_id FROM products WHERE id = ?"
	err = db.QueryRow(query, deleteReq.ID).Scan(&checkStoreID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}

	// Verify user owns the store
	var storeOwnerID int
	query = "SELECT id FROM stores WHERE owner_id = ? AND id = ?"
	err = db.QueryRow(query, userID, checkStoreID).Scan(&storeOwnerID)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized to delete this product"})
		return
	}

	// Delete product (CASCADE will handle related records)
	deleteQuery := "DELETE FROM products WHERE id = ?"
	result, err := db.Exec(deleteQuery, deleteReq.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete product"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
