package api

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
)

func AddToCart(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	productID := r.FormValue("product_id")
	quantity := r.FormValue("quantity")

	if productID == "" || quantity == "" {
		http.Error(w, `{"error": "Product ID and quantity required"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check if item already in cart
	var existingID int
	err = db.QueryRow("SELECT id FROM cart WHERE user_id = ? AND product_id = ?", userID, productID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new cart item
		_, err = db.Exec("INSERT INTO cart (user_id, product_id, quantity) VALUES (?, ?, ?)", userID, productID, quantity)
	} else {
		// Update existing cart item
		_, err = db.Exec("UPDATE cart SET quantity = quantity + ? WHERE id = ?", quantity, existingID)
	}

	if err != nil {
		http.Error(w, `{"error": "Failed to add to cart"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func GetCartTotal(w http.ResponseWriter, r *http.Request) {
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

	var totalItems int
	var totalPrice float64

	err = db.QueryRow(`
		SELECT COALESCE(SUM(c.quantity), 0), COALESCE(SUM(c.quantity * p.price), 0)
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ?
	`, userID).Scan(&totalItems, &totalPrice)

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, `{"error": "Failed to fetch cart"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_items": totalItems,
		"total_price": totalPrice,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetCart(w http.ResponseWriter, r *http.Request, store_id int) CartResponse {
	cartData, err := getCartData(w, r, store_id)
	if err != nil {
		return CartResponse{}
	}
	return cartData
}

func getCartData(w http.ResponseWriter, r *http.Request, store_id int) (CartResponse, error) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return CartResponse{}, nil
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return CartResponse{}, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.user_id, c.quantity,
			   p.id, p.name, p.description, p.price, p.image, p.store_id
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ? AND p.store_id = ?
	`, userID, store_id)

	if err != nil {
		return CartResponse{}, err
	}
	defer rows.Close()

	var items []CartItem
	totalPrice := 0.0
	totalItems := 0

	for rows.Next() {
		var id, productID, userID, quantity, storeID int
		var product Product

		err := rows.Scan(&id, &productID, &userID, &quantity,
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Image, &storeID)
		if err != nil {
			continue
		}

		itemTotal := product.Price * float64(quantity)
		totalPrice += itemTotal
		totalItems += quantity

		items = append(items, CartItem{
			ID:        id,
			ProductID: productID,
			Quantity:  quantity,
			Product:   product,
			ItemTotal: itemTotal,
		})
	}

	return CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}, nil
}

func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cartItemID := r.FormValue("cart_item_id")
	quantity := r.FormValue("quantity")

	if cartItemID == "" || quantity == "" {
		http.Error(w, `{"error": "Cart item ID and quantity required"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE cart SET quantity = ? WHERE id = ? AND user_id = ?", quantity, cartItemID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to update cart"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cartItemID := r.FormValue("cart_item_id")

	if cartItemID == "" {
		http.Error(w, `{"error": "Cart item ID required"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM cart WHERE id = ? AND user_id = ?", cartItemID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to remove from cart"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func CheckoutPage(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
	if !validUser {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer db.Close()

	// Get cart items
	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.user_id, c.quantity,
			   p.id, p.name, p.description, p.price, p.image, p.store_id
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ?
	`, userID)

	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to fetch cart")
		return
	}
	defer rows.Close()

	var items []CartItem
	totalPrice := 0.0
	totalItems := 0

	for rows.Next() {
		var id, productID, userID, quantity, storeID int
		var product Product

		err := rows.Scan(&id, &productID, &userID, &quantity,
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Image, &storeID)
		if err != nil {
			continue
		}

		itemTotal := product.Price * float64(quantity)
		totalPrice += itemTotal
		totalItems += quantity

		items = append(items, CartItem{
			ID:        id,
			ProductID: productID,
			Quantity:  quantity,
			Product:   product,
			ItemTotal: itemTotal,
		})
	}

	data := CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}

	tmpl, err := template.ParseFiles("../frontend/checkout.html")
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}

func CheckoutPageForStore(w http.ResponseWriter, r *http.Request, storeID int, store Store) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.user_id, c.quantity,
			   p.id, p.name, p.description, p.price, p.image, p.store_id
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ? AND p.store_id = ?
	`, userID, storeID)

	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to fetch cart")
		return
	}
	defer rows.Close()

	var items []CartItem
	totalPrice := 0.0
	totalItems := 0

	for rows.Next() {
		var id, productID, userID, quantity, storeID int
		var product Product

		err := rows.Scan(&id, &productID, &userID, &quantity,
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Image, &storeID)
		if err != nil {
			continue
		}

		itemTotal := product.Price * float64(quantity)
		totalPrice += itemTotal
		totalItems += quantity

		items = append(items, CartItem{
			ID:        id,
			ProductID: productID,
			Quantity:  quantity,
			Product:   product,
			ItemTotal: itemTotal,
		})
	}

	data := CartResponse{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
		Store:      store,
	}

	tmpl, err := template.ParseFiles("../frontend/checkout.html")
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}
