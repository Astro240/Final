package api

import (
	"database/sql"
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