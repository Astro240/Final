package api

import (
	"database/sql"
	"html/template"
	"log"
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
	tmpl, err := template.ParseFiles(FrontendStoreHTML)
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

	rows, err := db.Query("SELECT id, name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold FROM stores")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		var color Color
		var colorScheme string
		if err := rows.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID, &store.Phone, &store.Address, &store.PaymentMethods, &store.IBANNumber, &store.ShippingInfo, &store.ShippingCost, &store.EstimatedShipping, &store.FreeShippingThreshold); err != nil {
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

	rows, err := db.Query("SELECT id, name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold FROM stores WHERE owner_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		var color Color
		var colorScheme string
		if err := rows.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID, &store.Phone, &store.Address, &store.PaymentMethods, &store.IBANNumber, &store.ShippingInfo, &store.ShippingCost, &store.EstimatedShipping, &store.FreeShippingThreshold); err != nil {
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
	var color Color
	var colorScheme string
	err = db.QueryRow("SELECT id, name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold FROM stores WHERE id = ?", storeID).Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID, &store.Phone, &store.Address, &store.PaymentMethods, &store.IBANNumber, &store.ShippingInfo, &store.ShippingCost, &store.EstimatedShipping, &store.FreeShippingThreshold)
	if err != nil {
		if err == sql.ErrNoRows {
			return Store{}, nil
		}
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
	store.ColorScheme = color
	return store, nil
}

func CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateUser(w, r)
	if !validUser {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database connection failed"}`))
		return
	}
	defer db.Close()
	name := strings.TrimSpace(r.FormValue("storeTitle"))
	description := strings.TrimSpace(r.FormValue("storeDescription"))
	if err = ValidateStoreName(name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}
	query := "SELECT id FROM stores WHERE name = ?"
	row := db.QueryRow(query, name)
	var existingID int
	err = row.Scan(&existingID)
	if err != sql.ErrNoRows {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Store name already exists"}`))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Couldn't determine template colors"}`))
		return
	}

	logo := r.FormValue("storeLogo")
	avatarPath := StoreLogosPath
	logoImage := DefaultLogoFilename
	//check if logo is not empty, then validate and save
	if logo != "" {
		valid, filename := ValidateImage(avatarPath, logo)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "` + filename + `"}`))
			return
		}
		logoImage = filename
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Logo is required"}`))
		return
	}
	banner := r.FormValue("storeBanner")
	bannerPath := StoreBannersPath
	bannerImage := DefaultBannerFilename
	//check if banner is not empty, then validate and save
	if banner != "" {
		valid, filename := ValidateImage(bannerPath, banner)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "` + filename + `"}`))
			return
		}
		bannerImage = filename
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Banner is required"}`))
		return
	}
	colorscheme := strings.Join(colors, ",")

	// Get form values
	phone := strings.TrimSpace(r.FormValue("contactPhone"))
	address := strings.TrimSpace(r.FormValue("businessAddress"))
	paymentMethods := r.FormValue("paymentMethods") // This will be "credit_debit"
	ibanNumber := strings.TrimSpace(r.FormValue("ibanNumber"))
	shippingInfo := strings.TrimSpace(r.FormValue("shippingInfo"))
	shippingCostStr := r.FormValue("shippingCost")
	estimatedShippingStr := r.FormValue("estimatedShipping")
	freeShippingThresholdStr := r.FormValue("freeShippingThreshold")

	// Convert string values to appropriate types
	shippingCost := 0.0
	estimatedShipping := 0
	freeShippingThreshold := 0.0

	if shippingCostStr != "" {
		if val, err := parseFloat(shippingCostStr); err == nil {
			shippingCost = val
		}
	}

	if estimatedShippingStr != "" {
		if val, err := parseInt(estimatedShippingStr); err == nil {
			estimatedShipping = val
		}
	}

	if freeShippingThresholdStr != "" {
		if val, err := parseFloat(freeShippingThresholdStr); err == nil {
			freeShippingThreshold = val
		}
	}

	query = "INSERT INTO stores (name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(query, name, description, selectedTemplate, colorscheme, logoImage, bannerImage, userID, phone, address, paymentMethods, ibanNumber, shippingInfo, shippingCost, estimatedShipping, freeShippingThreshold)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to create store in database"}`))
		return
	}
	ip, err := GetIPv4()
	if err != nil {
		log.Println("Error getting IPv4:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to get server IP address"}`))
		return
	}
	err = AddHostEntry(strings.ToLower(name)+".com", ip)
	if err != nil {
		log.Println("Error adding host entry:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to configure store domain"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func GetStoreByName(storeName string) (Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return Store{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold FROM stores WHERE name = ?", storeName)
	if err != nil {
		return Store{}, err
	}
	var store Store
	var color Color
	var colorScheme string
	if err := row.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID, &store.Phone, &store.Address, &store.PaymentMethods, &store.IBANNumber, &store.ShippingInfo, &store.ShippingCost, &store.EstimatedShipping, &store.FreeShippingThreshold); err != nil {
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

// GetStoreByCustomDomain retrieves a store by its custom domain
func GetStoreByCustomDomain(customDomain string) (Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return Store{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, name, description, template, color_scheme, logo, banner, owner_id, phone, address, payment_methods, iban_number, shipping_info, shipping_cost, estimated_shipping, free_shipping_threshold FROM stores WHERE custom_domain = ?", customDomain)
	if err != nil {
		return Store{}, err
	}
	var store Store
	var color Color
	var colorScheme string
	if err := row.Scan(&store.ID, &store.Name, &store.Description, &store.Template, &colorScheme, &store.Logo, &store.Banner, &store.OwnerID, &store.Phone, &store.Address, &store.PaymentMethods, &store.IBANNumber, &store.ShippingInfo, &store.ShippingCost, &store.EstimatedShipping, &store.FreeShippingThreshold); err != nil {
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

// Helper function to parse float values
func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	i := 0
	sign := 1.0
	if len(s) > 0 && s[0] == '-' {
		sign = -1.0
		i = 1
	}

	num := 0.0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		num = num*10 + float64(s[i]-'0')
		i++
	}

	if i < len(s) && s[i] == '.' {
		i++
		dec := 0.1
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			num = num + dec*float64(s[i]-'0')
			dec *= 0.1
			i++
		}
	}

	return sign * num, nil
}

// Helper function to parse integer values
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	result := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		}
	}

	return result, nil
}
