package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// CreateReview allows users to review a product they have received
func CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, validUser := ValidateCustomer(w, r)
	if !validUser {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get form data
	productIDStr := strings.TrimSpace(r.FormValue("product_id"))
	ratingStr := strings.TrimSpace(r.FormValue("rating"))
	comment := strings.TrimSpace(r.FormValue("comment"))

	if productIDStr == "" || ratingStr == "" {
		http.Error(w, `{"error": "Product ID and rating are required"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid product ID"}`, http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, `{"error": "Rating must be between 1 and 5"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Find an eligible order for this user and product
	// Order must be shipped or completed, contain the product, and not already reviewed
	var orderID int
	err = db.QueryRow(`
		SELECT o.id 
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		WHERE o.user_id = ? 
		  AND op.product_id = ?
		  AND (o.status = 'shipped' OR o.status = 'completed')
		  AND NOT EXISTS (
			SELECT 1 FROM reviews r 
			WHERE r.order_id = o.id 
			  AND r.product_id = ? 
			  AND r.user_id = ?
		  )
		ORDER BY o.created_at DESC
		LIMIT 1
	`, userID, productID, productID, userID).Scan(&orderID)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "No eligible order found. You can only review products from shipped or completed orders that you haven't reviewed yet."}`, http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Failed to find eligible order"}`, http.StatusInternalServerError)
		return
	}

	// Insert the review
	result, err := db.Exec(`
		INSERT INTO reviews (product_id, user_id, order_id, rating, comment, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
	`, productID, userID, orderID, rating, comment)

	if err != nil {
		http.Error(w, `{"error": "Failed to create review"}`, http.StatusInternalServerError)
		return
	}

	reviewID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve review ID"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"review_id": reviewID,
		"message":   "Review submitted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetProductReviews retrieves all reviews for a specific product
func GetProductReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		http.Error(w, `{"error": "Product ID is required"}`, http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid product ID"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get all reviews for the product
	rows, err := db.Query(`
		SELECT r.id, r.product_id, r.user_id, r.order_id, r.rating, r.comment, r.created_at,
		       u.first_name || ' ' || COALESCE(u.last_name, '') as user_name
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = ?
		ORDER BY r.created_at DESC
	`, productID)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch reviews"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		err := rows.Scan(&review.ID, &review.ProductID, &review.UserID, &review.OrderID,
			&review.Rating, &review.Comment, &review.CreatedAt, &review.UserName)
		if err != nil {
			continue
		}
		reviews = append(reviews, review)
	}

	// Calculate average rating
	var avgRating float64
	var reviewCount int
	err = db.QueryRow(`
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM reviews
		WHERE product_id = ?
	`, productID).Scan(&avgRating, &reviewCount)

	if err != nil {
		http.Error(w, `{"error": "Failed to calculate average rating"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":        true,
		"reviews":        reviews,
		"average_rating": avgRating,
		"review_count":   reviewCount,
	}

	json.NewEncoder(w).Encode(response)
}

// GetUserReviewableOrders returns orders that the user can review products from
func GetUserReviewableOrders(w http.ResponseWriter, r *http.Request) {
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

	// Get orders with products that can be reviewed
	rows, err := db.Query(`
		SELECT DISTINCT o.id, o.total_amount, o.status, o.created_at,
		       op.product_id, p.name, p.image,
		       CASE WHEN r.id IS NOT NULL THEN 1 ELSE 0 END as has_reviewed
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		JOIN products p ON op.product_id = p.id
		LEFT JOIN reviews r ON r.order_id = o.id AND r.product_id = op.product_id AND r.user_id = ?
		WHERE o.user_id = ? AND (o.status = 'shipped' OR o.status = 'completed')
		ORDER BY o.created_at DESC
	`, userID, userID)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch reviewable orders"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ReviewableProduct struct {
		ProductID    int    `json:"product_id"`
		ProductName  string `json:"product_name"`
		ProductImage string `json:"product_image"`
		HasReviewed  bool   `json:"has_reviewed"`
	}

	type ReviewableOrder struct {
		OrderID     int                 `json:"order_id"`
		TotalAmount float64             `json:"total_amount"`
		Status      string              `json:"status"`
		CreatedAt   string              `json:"created_at"`
		Products    []ReviewableProduct `json:"products"`
	}

	ordersMap := make(map[int]*ReviewableOrder)

	for rows.Next() {
		var orderID, productID int
		var totalAmount float64
		var status, createdAt, productName, productImage string
		var hasReviewed bool

		err := rows.Scan(&orderID, &totalAmount, &status, &createdAt, &productID, &productName, &productImage, &hasReviewed)
		if err != nil {
			continue
		}

		if _, exists := ordersMap[orderID]; !exists {
			ordersMap[orderID] = &ReviewableOrder{
				OrderID:     orderID,
				TotalAmount: totalAmount,
				Status:      status,
				CreatedAt:   createdAt,
				Products:    []ReviewableProduct{},
			}
		}

		ordersMap[orderID].Products = append(ordersMap[orderID].Products, ReviewableProduct{
			ProductID:    productID,
			ProductName:  productName,
			ProductImage: productImage,
			HasReviewed:  hasReviewed,
		})
	}

	orders := make([]ReviewableOrder, 0, len(ordersMap))
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	response := map[string]interface{}{
		"success": true,
		"orders":  orders,
	}

	json.NewEncoder(w).Encode(response)
}
