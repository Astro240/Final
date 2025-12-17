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
		valid, filename := ValidateImage(avatarPath, logo)
		if !valid {
			http.Error(w, `{"error": "`+filename+`"}`, http.StatusInternalServerError)
			return
		}
		logoImage = filename
	} else {
		http.Error(w, `{"error": "Logo is required"}`, http.StatusInternalServerError)
		return
	}
	banner := r.FormValue("storeBanner")
	bannerPath := "./store_images/banners/"
	bannerImage := "default.png"
	//check if banner is not empty, then validate and save
	if banner != "" {
		valid, filename := ValidateImage(bannerPath, banner)
		if !valid {
			http.Error(w, `{"error": "`+filename+`"}`, http.StatusInternalServerError)
			return
		}
		bannerImage = filename
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
	ip, err := GetIPv4()
	if err != nil {
		log.Fatal(err)
	}
	err = AddHostEntry(strings.ToLower(name)+".com", ip)
	if err != nil {
		log.Fatal(err)
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

// GetStoreByCustomDomain retrieves a store by its custom domain
func GetStoreByCustomDomain(customDomain string) (Store, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return Store{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, name, description, template, color_scheme, logo, banner, owner_id FROM stores WHERE custom_domain = ?", customDomain)
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
