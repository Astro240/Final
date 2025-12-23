package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	// Parse quantity
	var requestedQty int
	_, err = fmt.Sscanf(quantity, "%d", &requestedQty)
	if err != nil || requestedQty <= 0 {
		http.Error(w, `{"error": "Invalid quantity"}`, http.StatusBadRequest)
		return
	}

	// Check product's available quantity
	var availableQty int
	err = db.QueryRow("SELECT quantity FROM products WHERE id = ?", productID).Scan(&availableQty)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Product not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to check product availability"}`, http.StatusInternalServerError)
		return
	}

	// Check if item already in cart and get current cart quantity
	var existingID int
	var currentCartQty int
	err = db.QueryRow("SELECT id, quantity FROM cart WHERE user_id = ? AND product_id = ?", userID, productID).Scan(&existingID, &currentCartQty)

	var newTotalQty int
	if err == sql.ErrNoRows {
		// New item - check if requested quantity is available
		newTotalQty = requestedQty
		if newTotalQty > availableQty {
			http.Error(w, fmt.Sprintf(`{"error": "Only %d items available"}`, availableQty), http.StatusBadRequest)
			return
		}
		// Insert new cart item
		_, err = db.Exec("INSERT INTO cart (user_id, product_id, quantity) VALUES (?, ?, ?)", userID, productID, requestedQty)
		if err != nil {
			http.Error(w, `{"error": "Failed to add to cart"}`, http.StatusInternalServerError)
			return
		}
		// Decrease product quantity
		_, err = db.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?", requestedQty, productID)
	} else if err == nil {
		// Update existing cart item - check if adding would exceed available quantity
		newTotalQty = currentCartQty + requestedQty
		if newTotalQty > availableQty {
			http.Error(w, fmt.Sprintf(`{"error": "Only %d items available, you already have %d in cart"}`, availableQty, currentCartQty), http.StatusBadRequest)
			return
		}
		_, err = db.Exec("UPDATE cart SET quantity = ? WHERE id = ?", newTotalQty, existingID)
		if err != nil {
			http.Error(w, `{"error": "Failed to update cart"}`, http.StatusInternalServerError)
			return
		}
		// Decrease product quantity by the additional amount
		_, err = db.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?", requestedQty, productID)
	} else {
		http.Error(w, `{"error": "Failed to check cart"}`, http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, `{"error": "Failed to update product quantity"}`, http.StatusInternalServerError)
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

	// Parse new quantity
	var newQty int
	_, err = fmt.Sscanf(quantity, "%d", &newQty)
	if err != nil || newQty <= 0 {
		http.Error(w, `{"error": "Invalid quantity"}`, http.StatusBadRequest)
		return
	}

	// Get current cart item details
	var currentQty, productID int
	err = db.QueryRow("SELECT quantity, product_id FROM cart WHERE id = ? AND user_id = ?", cartItemID, userID).Scan(&currentQty, &productID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Cart item not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch cart item"}`, http.StatusInternalServerError)
		return
	}

	// Calculate the difference
	qtyDifference := newQty - currentQty

	if qtyDifference > 0 {
		// User wants to add more - check if available
		var availableQty int
		err = db.QueryRow("SELECT quantity FROM products WHERE id = ?", productID).Scan(&availableQty)
		if err != nil {
			http.Error(w, `{"error": "Failed to check product availability"}`, http.StatusInternalServerError)
			return
		}

		if qtyDifference > availableQty {
			http.Error(w, fmt.Sprintf(`{"error": "Only %d more items available"}`, availableQty), http.StatusBadRequest)
			return
		}

		// Decrease product quantity
		_, err = db.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?", qtyDifference, productID)
		if err != nil {
			http.Error(w, `{"error": "Failed to update product quantity"}`, http.StatusInternalServerError)
			return
		}
	} else if qtyDifference < 0 {
		// User wants to decrease - restore product quantity
		restoreQty := -qtyDifference
		_, err = db.Exec("UPDATE products SET quantity = quantity + ? WHERE id = ?", restoreQty, productID)
		if err != nil {
			http.Error(w, `{"error": "Failed to restore product quantity"}`, http.StatusInternalServerError)
			return
		}
	}

	// Update cart
	_, err = db.Exec("UPDATE cart SET quantity = ? WHERE id = ? AND user_id = ?", newQty, cartItemID, userID)
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

	// Get the cart item details before deleting (to restore product quantity)
	var productID int
	var quantity int
	err = db.QueryRow("SELECT product_id, quantity FROM cart WHERE id = ? AND user_id = ?", cartItemID, userID).Scan(&productID, &quantity)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Cart item not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch cart item"}`, http.StatusInternalServerError)
		return
	}

	// Delete from cart
	_, err = db.Exec("DELETE FROM cart WHERE id = ? AND user_id = ?", cartItemID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to remove from cart"}`, http.StatusInternalServerError)
		return
	}

	// Restore product quantity
	_, err = db.Exec("UPDATE products SET quantity = quantity + ? WHERE id = ?", quantity, productID)
	if err != nil {
		http.Error(w, `{"error": "Failed to restore product quantity"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func CheckoutPageForStore(w http.ResponseWriter, r *http.Request, storeID int, store Store) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		HandleError(w, r, http.StatusUnauthorized, "Unauthorized to view, please login")
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

	tmpl, err := template.ParseFiles(FrontendCheckoutHTML)
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}
