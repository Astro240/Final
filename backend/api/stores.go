package api

import (
	"database/sql"
	"html/template"
	"net/http"

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

	rows, err := db.Query("SELECT id, name,description FROM stores")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		if err := rows.Scan(&store.ID, &store.Name, &store.Description); err != nil {
			return nil, err
		}
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
	
	rows, err := db.Query("SELECT id, name,description FROM stores WHERE owner_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var store Store
		if err := rows.Scan(&store.ID, &store.Name, &store.Description); err != nil {
			return nil, err
		}
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
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	name := r.FormValue("name")
	description := r.FormValue("description")
	colorScheme := r.FormValue("color_scheme")
	logo := r.FormValue("logo")
	avatarPath := "./avatars/"
	logoImage := "default.png"
	//check if logo is not empty, then validate and save
	if logo != "" {
		valid, logoImage := ValidateImage(avatarPath, logo)
		if !valid {
			http.Error(w, `{"error": "`+logoImage+`"}`, http.StatusInternalServerError)
			return
		}
	}
	query := "INSERT INTO stores (name, description, color_scheme, logo, owner_id) VALUES (?, ?, ?, ?, ?)"
	_, err = db.Exec(query, name, description, colorScheme, logoImage, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
