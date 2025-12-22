package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DashboardStats struct {
	TotalRevenue   float64              `json:"total_revenue"`
	ItemsSold      int                  `json:"items_sold"`
	ActiveProducts int                  `json:"active_products"`
	AvgOrderValue  float64              `json:"avg_order_value"`
	RevenueChange  float64              `json:"revenue_change"`
	SalesChange    float64              `json:"sales_change"`
	ProductsChange float64              `json:"products_change"`
	AvgChange      float64              `json:"avg_change"`
	SalesData      []SalesDataPoint     `json:"sales_data"`
	TopProducts    []ProductPerformance `json:"top_products"`
}

type SalesDataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type ProductPerformance struct {
	DateAdded   string  `json:"date_added"`
	ProductName string  `json:"product_name"`
	Performance string  `json:"performance"`
	Category    string  `json:"category"`
	UnitsSold   int     `json:"units_sold"`
	Price       float64 `json:"price"`
	Revenue     float64 `json:"revenue"`
	Profit      float64 `json:"profit"`
	Margin      float64 `json:"margin"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// Handle incoming messages
	}
}

func DashboardWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Get store ID from query parameters
	storeIDStr := r.URL.Query().Get("store_id")
	if storeIDStr == "" {
		return
	}

	storeID, err := strconv.Atoi(storeIDStr)
	if err != nil {
		return
	}

	// Verify user owns the store
	user, err := GetCookie(r, "session_token")
	if err != nil {
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return
	}
	defer db.Close()

	// Verify ownership
	var ownerID int
	err = db.QueryRow("SELECT owner_id FROM stores WHERE id = ?", storeID).Scan(&ownerID)
	if err != nil {
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_token = ?", user).Scan(&userID)
	if err != nil || userID != ownerID {
		return
	}

	// Fetch store and set cookie
	store, err := GetStoreByID(storeID)
	if err == nil && store.ID != 0 {
		SetStoreCookie(w, int(store.OwnerID), store.Name)
	}

	// Configure ping/pong for connection health check
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Send initial data
	stats := getDashboardStats(db, storeID)

	if err := conn.WriteJSON(stats); err != nil {
		return
	}

	// Channel to signal when to stop
	done := make(chan struct{})

	// Goroutine to read messages and detect disconnections
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Send updates every 5 seconds and ping every 30 seconds
	ticker := time.NewTicker(5 * time.Second)
	pingTicker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer pingTicker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			stats := getDashboardStats(db, storeID)
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(stats); err != nil {
				return
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func getDashboardStats(db *sql.DB, storeID int) DashboardStats {
	stats := DashboardStats{
		SalesData:   make([]SalesDataPoint, 0),
		TopProducts: make([]ProductPerformance, 0),
	}

	// Get total revenue and items sold from orders
	var totalRevenue sql.NullFloat64
	var itemsSold sql.NullInt64
	query := `
		SELECT 
			COALESCE(SUM(op.quantity * op.price), 0) as total_revenue,
			COALESCE(SUM(op.quantity), 0) as items_sold
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		JOIN products p ON op.product_id = p.id
		WHERE p.store_id = ? AND (o.status = 'paid' OR o.status= 'shipped' 
		OR o.status = 'completed')
	`
	err := db.QueryRow(query, storeID).Scan(&totalRevenue, &itemsSold)

	if totalRevenue.Valid {
		stats.TotalRevenue = totalRevenue.Float64
	}
	if itemsSold.Valid {
		stats.ItemsSold = int(itemsSold.Int64)
	}

	// Get active products count
	var activeProducts int
	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE store_id = ?", storeID).Scan(&activeProducts)

	stats.ActiveProducts = activeProducts

	// Calculate average order value
	if stats.ItemsSold > 0 {
		stats.AvgOrderValue = stats.TotalRevenue / float64(stats.ItemsSold)
	}

	// Calculate percentage changes based on last month vs current month
	currentMonth := time.Now().Month()
	currentYear := time.Now().Year()
	lastMonth := currentMonth - 1
	lastYear := currentYear
	if lastMonth == 0 {
		lastMonth = 12
		lastYear = currentYear - 1
	}

	// Get last month's revenue
	var lastMonthRevenue sql.NullFloat64
	var lastMonthSales sql.NullInt64
	lastMonthQuery := `
		SELECT 
			COALESCE(SUM(op.quantity * op.price), 0) as revenue,
			COALESCE(SUM(op.quantity), 0) as items_sold
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		JOIN products p ON op.product_id = p.id
		WHERE p.store_id = ? 
		AND strftime('%m', o.created_at) = ?
		AND strftime('%Y', o.created_at) = ?
		AND (o.status = 'paid' OR o.status = 'shipped' OR o.status = 'completed')
	`
	db.QueryRow(lastMonthQuery, storeID, fmt.Sprintf("%02d", lastMonth), fmt.Sprintf("%d", lastYear)).
		Scan(&lastMonthRevenue, &lastMonthSales)

	// Calculate revenue change percentage
	if lastMonthRevenue.Valid && lastMonthRevenue.Float64 > 0 {
		stats.RevenueChange = ((stats.TotalRevenue - lastMonthRevenue.Float64) / lastMonthRevenue.Float64) * 100
	} else if stats.TotalRevenue > 0 {
		// New business with no last month data but has revenue this month = 100% growth
		stats.RevenueChange = 100
	} else {
		stats.RevenueChange = 0
	}

	// Calculate sales change percentage
	if lastMonthSales.Valid && lastMonthSales.Int64 > 0 {
		stats.SalesChange = ((float64(stats.ItemsSold) - float64(lastMonthSales.Int64)) / float64(lastMonthSales.Int64)) * 100
	} else if stats.ItemsSold > 0 {
		// New business with no last month data but has sales this month = 100% growth
		stats.SalesChange = 100
	} else {
		stats.SalesChange = 0
	}

	// Get last month's average order value for comparison
	var lastMonthAvgOrderValue float64
	if lastMonthSales.Valid && lastMonthSales.Int64 > 0 && lastMonthRevenue.Valid {
		lastMonthAvgOrderValue = lastMonthRevenue.Float64 / float64(lastMonthSales.Int64)
	}

	// Calculate average order value change percentage
	if lastMonthAvgOrderValue > 0 {
		stats.AvgChange = ((stats.AvgOrderValue - lastMonthAvgOrderValue) / lastMonthAvgOrderValue) * 100
	} else if stats.AvgOrderValue > 0 {
		// New business with no last month data but has avg order value = 100% growth
		stats.AvgChange = 100
	} else {
		stats.AvgChange = 0
	}

	// Get last month's product count for comparison
	var lastMonthProductCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM products 
		WHERE store_id = ? 
		AND created_at < date('now', 'start of month')
	`, storeID).Scan(&lastMonthProductCount)

	// Calculate products change percentage
	if lastMonthProductCount > 0 {
		stats.ProductsChange = ((float64(stats.ActiveProducts) - float64(lastMonthProductCount)) / float64(lastMonthProductCount)) * 100
	} else if stats.ActiveProducts > 0 {
		// New business with no last month data but has products = 100% growth
		stats.ProductsChange = 100
	} else {
		stats.ProductsChange = 0
	}

	// Get monthly sales data
	monthlyQuery := `
		SELECT 
			strftime('%m', o.created_at) as month,
			COALESCE(SUM(op.quantity * op.price), 0) as revenue
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		JOIN products p ON op.product_id = p.id
		WHERE p.store_id = ? AND (o.status = 'paid' OR o.status= 'shipped' 
		OR o.status = 'completed')
		GROUP BY month
		ORDER BY month
	`
	rows, err := db.Query(monthlyQuery, storeID)
	if err == nil {
		defer rows.Close()
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		for rows.Next() {
			var month string
			var revenue float64
			if err := rows.Scan(&month, &revenue); err == nil {
				monthInt, _ := strconv.Atoi(month)
				if monthInt > 0 && monthInt <= 12 {
					stats.SalesData = append(stats.SalesData, SalesDataPoint{
						Label: months[monthInt-1],
						Value: revenue,
					})
				}
			}
		}
	}

	// Get top performing products (up to 5)
	// First, get products with actual sales
	topProductsQuery := `
		SELECT 
			p.name,
			p.price,
			COALESCE(SUM(op.quantity), 0) as units_sold,
			COALESCE(SUM(op.quantity * op.price), 0) as revenue,
			p.created_at
		FROM products p
		LEFT JOIN order_products op ON p.id = op.product_id
		WHERE p.store_id = ?
		GROUP BY p.id
		HAVING units_sold > 0
		ORDER BY revenue DESC, units_sold DESC
		LIMIT 5
	`
	rows, err = db.Query(topProductsQuery, storeID)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var perf ProductPerformance
			var unitsSold int
			var revenue float64
			var createdAt string
			if err := rows.Scan(&perf.ProductName, &perf.Price, &unitsSold, &revenue, &createdAt); err == nil {
				perf.UnitsSold = unitsSold
				perf.Revenue = revenue
				perf.Profit = revenue * 0.3 // 30% profit margin assumption
				perf.Margin = 30.0
				perf.Category = "General"

				// Parse and format the created_at date
				if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
					perf.DateAdded = t.Format("01/02/2006")
				} else {
					perf.DateAdded = time.Now().Format("01/02/2006")
				}

				// Determine performance based on units sold
				if unitsSold > 30 {
					perf.Performance = "HOT"
				} else if unitsSold > 10 {
					perf.Performance = "WARM"
				} else {
					perf.Performance = "SLOW"
				}

				stats.TopProducts = append(stats.TopProducts, perf)
			}
		}
	}

	// If we have fewer than 5 products with sales, fill with products without sales
	if len(stats.TopProducts) < 5 {
		remainingQuery := `
			SELECT 
				p.name,
				p.price,
				p.created_at
			FROM products p
			WHERE p.store_id = ? 
			AND p.id NOT IN (
				SELECT DISTINCT product_id FROM order_products
			)
			ORDER BY p.created_at DESC
			LIMIT ?
		`
		limit := 5 - len(stats.TopProducts)
		rows, err = db.Query(remainingQuery, storeID, limit)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var perf ProductPerformance
				var createdAt string
				if err := rows.Scan(&perf.ProductName, &perf.Price, &createdAt); err == nil {
					perf.UnitsSold = 0
					perf.Revenue = 0
					perf.Profit = 0
					perf.Margin = 0
					perf.Category = "General"
					perf.Performance = "NEW"

					if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
						perf.DateAdded = t.Format("01/02/2006")
					} else {
						perf.DateAdded = time.Now().Format("01/02/2006")
					}

					stats.TopProducts = append(stats.TopProducts, perf)
				}
			}
		}
	}

	return stats
}
