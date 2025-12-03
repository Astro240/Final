package api

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func StorePage(w http.ResponseWriter, r *http.Request) {
	store, err := GetStores()
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Unable to retrieve stores")
		return
	}
	var myStore []Store
	userID, validUser := ValidateUser(w, r)
	if validUser {
		myStore, err = GetMyStores(userID)
		if err != nil {
			HandleError(w, r, http.StatusInternalServerError, "Unable to retrieve your stores")
			return
		}
	}
	tmpl, err := template.ParseFiles("../frontend/store.html")
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}
	Stores := StoreDisplay{
		MyStores:  myStore,
		AllStores: store,
	}
	if err := tmpl.Execute(w, Stores); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}

func GetStores() ([]Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name,description, template, color_scheme, logo, banner, owner_id FROM stores")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		var color Color
		var colorScheme string
		if err := rows.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID); err != nil {
			return nil, err
		}
		colors := strings.Split(colorScheme, ",")
		for i := 0; i < len(colors); i++ {
			if i == 0 {
				color.Primary = colors[i]
			} else if i == 1 {
				color.Secondary = colors[i]
			} else if i == 2 {
				color.Background = colors[i]
			} else if i == 3 {
				color.Accent = colors[i]
			} else if i == 4 {
				color.Supporting = colors[i]
			} else if i == 5 {
				color.Tertiary = colors[i]
			} else if i == 6 {
				color.Highlight = colors[i]
			}
		}
		store.ColorScheme = color
		stores = append(stores, store)
	}
	return stores, nil
}

func GetMyStores(userID int) ([]Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name,description, template, color_scheme, logo, banner, owner_id FROM stores WHERE owner_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		var color Color
		var colorScheme string
		if err := rows.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID); err != nil {
			return nil, err
		}
		colors := strings.Split(colorScheme, ",")
		for i := 0; i < len(colors); i++ {
			if i == 0 {
				color.Primary = colors[i]
			} else if i == 1 {
				color.Secondary = colors[i]
			} else if i == 2 {
				color.Background = colors[i]
			} else if i == 3 {
				color.Accent = colors[i]
			} else if i == 4 {
				color.Supporting = colors[i]
			} else if i == 5 {
				color.Tertiary = colors[i]
			} else if i == 6 {
				color.Highlight = colors[i]
			}
		}
		store.ColorScheme = color
		stores = append(stores, store)
	}
	return stores, nil
}

func GetStoreByID(storeID int) (Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return Store{}, err
	}
	defer db.Close()

	var store Store
	err = db.QueryRow("SELECT id, name FROM stores WHERE id = ?", storeID).Scan(&store.ID, &store.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Store{}, nil
		}
		return Store{}, err
	}
	return store, nil
}

func CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
	if !validUser {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	name := strings.TrimSpace(r.FormValue("storeTitle"))
	description := strings.TrimSpace(r.FormValue("storeDescription"))
	if err = ValidateStoreName(name); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	query := "SELECT id FROM stores WHERE name = ?"
	row := db.QueryRow(query, name)
	var existingID int
	err = row.Scan(&existingID)
	if err != sql.ErrNoRows {
		http.Error(w, `{"error": "Store name already exists"}`, http.StatusInternalServerError)
		return
	}

	selectedTemplate := r.FormValue("selectedTemplate")
	colors := []string{}
	if selectedTemplate == "modern" {
		colors = append(colors,
			r.FormValue("modern-primary"),
			r.FormValue("modern-secondary"),
			r.FormValue("modern-background"),
			r.FormValue("modern-accent"))
	} else if selectedTemplate == "vibrant" {
		colors = append(colors,
			r.FormValue("vibrant-primary"),
			r.FormValue("vibrant-secondary"),
			r.FormValue("vibrant-background"),
			r.FormValue("vibrant-accent"),
			r.FormValue("vibrant-tertiary"),
			r.FormValue("vibrant-supporting"),
			r.FormValue("vibrant-highlight"))
	} else if selectedTemplate == "luxury" {
		colors = append(colors,
			r.FormValue("luxury-primary"),
			r.FormValue("luxury-secondary"),
			r.FormValue("luxury-background"))
	} else {
		http.Error(w, `{"error": "Couldn't determine template colors"}`, http.StatusInternalServerError)
		return
	}

	logo := r.FormValue("storeLogo")
	avatarPath := "./store_images/logos/"
	logoImage := "default.png"
	//check if logo is not empty, then validate and save
	if logo != "" {
		valid, logoImage := ValidateImage(avatarPath, logo)
		if !valid {
			http.Error(w, `{"error": "`+logoImage+`"}`, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, `{"error": "Logo is required"}`, http.StatusInternalServerError)
		return
	}
	banner := r.FormValue("storeBanner")
	bannerPath := "./store_images/banners/"
	bannerImage := "default.png"
	//check if banner is not empty, then validate and save
	if banner != "" {
		valid, bannerImage := ValidateImage(bannerPath, banner)
		if !valid {
			http.Error(w, `{"error": "`+bannerImage+`"}`, http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, `{"error": "Banner is required"}`, http.StatusInternalServerError)
		return
	}
	colorscheme := strings.Join(colors, ",")
	query = "INSERT INTO stores (name, description,template, color_scheme, logo, banner, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(query, name, description, selectedTemplate, colorscheme, logoImage, bannerImage, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetStoreByName(storeName string) (Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return Store{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, name,description, template, color_scheme, logo, banner, owner_id FROM stores WHERE name = ?", storeName)
	if err != nil {
		return Store{}, err
	}
	var store Store
	var color Color
	var colorScheme string
	if err := row.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID); err != nil {
		return Store{}, err
	}
	colors := strings.Split(colorScheme, ",")
	for i := 0; i < len(colors); i++ {
		if i == 0 {
			color.Primary = colors[i]
		} else if i == 1 {
			color.Secondary = colors[i]
		} else if i == 2 {
			color.Background = colors[i]
		} else if i == 3 {
			color.Accent = colors[i]
		} else if i == 4 {
			color.Supporting = colors[i]
		} else if i == 5 {
			color.Tertiary = colors[i]
		} else if i == 6 {
			color.Highlight = colors[i]
		}
	}
	store.Products, err = GetProductsByStoreID(store.ID)
	if err != nil {
		return Store{}, err
	}
	store.ColorScheme = color
	return store, nil
}

// Cart and Checkout related functions

func AddToCart(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
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

func GetCart(w http.ResponseWriter, r *http.Request) {
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
		SELECT c.id, c.product_id, c.user_id, c.quantity,
			   p.id, p.name, p.description, p.price, p.image, p.store_id
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ?
	`, userID)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch cart"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	totalPrice := 0.0
	totalItems := 0

	for rows.Next() {
		var id, productID, userID, quantity, pID, storeID int
		var name, description, image string
		var price float64

		err := rows.Scan(&id, &productID, &userID, &quantity, &pID, &name, &description, &price, &image, &storeID)
		if err != nil {
			continue
		}

		itemTotal := price * float64(quantity)
		totalPrice += itemTotal
		totalItems += quantity

		items = append(items, map[string]interface{}{
			"id":         id,
			"product_id": productID,
			"quantity":   quantity,
			"product": map[string]interface{}{
				"id":          pID,
				"name":        name,
				"description": description,
				"price":       price,
				"image":       image,
				"store_id":    storeID,
			},
			"item_total": itemTotal,
		})
	}

	response := map[string]interface{}{
		"items":       items,
		"total_items": totalItems,
		"total_price": totalPrice,
	}

	json.NewEncoder(w).Encode(response)
}

func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
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
	userID, validUser := ValidateUser(w, r)
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

	type CartItemDisplay struct {
		ID        int
		ProductID int
		Quantity  int
		Product   Product
		ItemTotal float64
	}

	type CheckoutDisplay struct {
		Items      []CartItemDisplay
		TotalItems int
		TotalPrice float64
	}

	var items []CartItemDisplay
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

		items = append(items, CartItemDisplay{
			ID:        id,
			ProductID: productID,
			Quantity:  quantity,
			Product:   product,
			ItemTotal: itemTotal,
		})
	}

	data := CheckoutDisplay{
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

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
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
