package api

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
)

func StorePage(w http.ResponseWriter, r *http.Request) {
	store, err := GetStores()
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Unable to retrieve stores")
		return
	}

	tmpl, err := template.ParseFiles("../frontend/store.html")
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}
	Stores := StoreDisplay{
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

func CreateStore(name string, ownerID int) (int64, error) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return 0, err
	}
	db.Close()
	query := "INSERT INTO stores (name, owner_id) VALUES (?, ?)"
	result, err := db.Exec(query, name, ownerID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
