package api

import (
	"database/sql"
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
	var User UserProfile
	userID, validUser := ValidateUser(w, r)
	if validUser {
		myStore, err = GetMyStores(userID)
		if err != nil {
			HandleError(w, r, http.StatusInternalServerError, "Unable to retrieve your stores")
			return
		}
		placeHolder, valid := getUser(userID)
		if valid {
			User = placeHolder
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
		User:      User,
		ValidUser: validUser,
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

func CheckoutPageForStore(w http.ResponseWriter, r *http.Request, storeID int,store Store) {
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
