package api

import (
	"database/sql"
	"net/http"
	"strconv"
)

func GetProductsByStoreID(storeID uint) ([]Product, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, name, description, price, image, quantity FROM items WHERE store_id = ?", storeID)
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
	w.Header().Set("Content-Type", "application/json")
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
	valid, imageName := ValidateImage("./store_images/products/", itemImage)
	if !valid {
		http.Error(w, `{"error": "`+imageName+`"}`, http.StatusInternalServerError)
		return
	}
	insertQuery := "INSERT INTO items (name, description, price, image, quantity, store_id) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(insertQuery, itemName, itemDescription, itemPrice, imageName, itemQuantity, storeOwnerID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create product"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"success": true}`))
}
