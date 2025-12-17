package api

import (
	"html/template"
	"net/http"
	"strings"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	storeName := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/"))
	isCheckout := false
	isDashboard := false
	isPayment := false
	isOrders := false
	// Check for dashboard/checkout/payment at the start of path
	if strings.HasPrefix(storeName, "checkout/payment") {
		isPayment = true
		storeName = ""
	} else if strings.HasPrefix(storeName, "checkout") {
		isCheckout = true
		storeName = ""
	} else if strings.HasPrefix(storeName, "dashboard") {
		isDashboard = true
		storeName = ""
	}

	// Check for orders suffix anywhere in the path (e.g., /storename/orders)
	if strings.HasSuffix(storeName, "/orders") && !isCheckout && !isDashboard && !isPayment {
		isOrders = true
		storeName = strings.TrimSuffix(storeName, "/orders")
	} else if storeName == "orders" {
		// For just /orders
		isOrders = true
		storeName = ""
	}

	// Check if path contains a custom domain like /(name).com
	if storeName != "" && strings.Contains(storeName, ".com") {
		// Extract store name from path like "mystore.com" or "mystore.com/checkout"
		pathParts := strings.Split(storeName, "/")
		domainPart := pathParts[0] // Get "mystore.com"

		// Check for checkout/dashboard/payment in remaining path
		remainingPath := strings.Join(pathParts[1:], "/")
		if strings.HasSuffix(remainingPath, "checkout/payment") {
			isPayment = true
		} else if strings.HasSuffix(remainingPath, "checkout") {
			isCheckout = true
		} else if strings.HasSuffix(remainingPath, "dashboard") {
			isDashboard = true
		} else if strings.HasPrefix(remainingPath, "orders") {
			isOrders = true
		}

		// Extract store name from domain (e.g., "mystore.com" -> "mystore")
		storeName = strings.TrimSuffix(domainPart, ".com")
	} else if storeName == "" {
		// Get hostname - try r.Host first (includes port), then r.URL.Hostname()
		hostname := r.Host

		if hostname == "" {
			hostname = r.URL.Hostname()
		}

		// Remove port if present
		if idx := strings.Index(hostname, ":"); idx != -1 {
			hostname = hostname[:idx]
		}
		// If hostname is not the main domain, treat it as a custom store domain
		if hostname != "" && hostname != "astropify.com" && hostname != "localhost" {
			storeName = extractStoreNameFromDomain(hostname)
		}
	} else {
		// Handle old path format like /mystore/checkout
		if strings.HasSuffix(storeName, "/checkout/payment") {
			isPayment = true
			storeName = strings.TrimSuffix(storeName, "/checkout/payment")
		} else if strings.HasSuffix(storeName, "/checkout") {
			isCheckout = true
			storeName = strings.TrimSuffix(storeName, "/checkout")
		} else if strings.HasSuffix(storeName, "/dashboard") {
			isDashboard = true
			storeName = strings.TrimSuffix(storeName, "/dashboard")
		}
		storeName = strings.TrimSuffix(storeName, ".com")
	}
	if storeName == "" {
		userID, validUser := ValidateUser(w, r)
		var User UserProfile
		if validUser {
			placeHolder, valid := getUser(userID)
			if valid {
				User = placeHolder
			}
		}
		tmpl, err := template.ParseFiles("../frontend/index.html")
		if err != nil {
			HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
			return
		}
		if err := tmpl.Execute(w, User); err != nil {
			HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
			return
		}
		return
	}
	if storeName != "" {
		store, err := GetStoreByName(storeName)
		if err == nil && store.ID != 0 && !isCheckout && !isDashboard && !isPayment && !isOrders {
			tmpl, err := template.ParseFiles("../frontend/templates/" + store.Template + "_template.html")
			if err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
				return
			}
			userID, validUser := ValidateUser(w, r)
			if validUser && uint(userID) == store.OwnerID {
				store.IsOwner = true
			}
			if err := tmpl.Execute(w, store); err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
				return
			}
			return
		} else if isCheckout && err == nil && store.ID != 0 {
			//Redirect to the proper checkout handler
			userID, validUser := ValidateCustomer(w, r)
			if !validUser || uint(userID) != store.OwnerID {
				HandleError(w, r, http.StatusForbidden, "Access denied")
				return
			}
			CheckoutPageForStore(w, r, int(store.ID), store)
			return
		} else if isDashboard && err == nil && store.ID != 0 {
			//validate the user and display the dashboard
			userID, validUser := ValidateUser(w, r)
			if !validUser || uint(userID) != store.OwnerID {
				HandleError(w, r, http.StatusForbidden, "Access denied")
				return
			}

			tmpl, err := template.ParseFiles("../frontend/dashboard.html")
			if err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
				return
			}

			dashboardData := DashboardData{
				Store: store,
			}

			if err := tmpl.Execute(w, dashboardData); err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
				return
			}
			return
		} else if isPayment && err == nil && store.ID != 0 {
			if _, validUser := ValidateCustomer(w, r); !validUser {
				HandleError(w, r, http.StatusUnauthorized, "Unauthorized, Please login!")
				return
			}

			tmpl, err := template.ParseFiles("../frontend/payment.html")
			if err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
				return
			}

			paymentData := PaymentPageData{
				Store: store,
			}

			if err := tmpl.Execute(w, paymentData); err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
				return
			}
			return
		} else if isOrders && err == nil && store.ID != 0 {
			// Show customer's orders for this store
			if _, validUser := ValidateCustomer(w, r); !validUser {
				HandleError(w, r, http.StatusUnauthorized, "Unauthorized, Please login!")
				return
			}

			tmpl, err := template.ParseFiles("../frontend/orders.html")
			if err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
				return
			}

			ordersData := map[string]interface{}{
				"Store": store,
			}

			if err := tmpl.Execute(w, ordersData); err != nil {
				HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
				return
			}
			return
		}
	}
	HandleError(w, r, http.StatusNotFound, "Invalid Path")
}

func SampleStoreView(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/templates/preview/"):]
	tmpl, err := template.ParseFiles("../frontend/templates/" + path + ".html")
	if err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to load template")
		return
	}
	sampleStore := Store{}
	if path == "modern_template" {
		sampleStore = Store{
			ID:       0,
			Name:     "Modern Store",
			Template: path,
			ColorScheme: Color{
				Primary:    "#0f172a",
				Secondary:  "#475569",
				Background: "#f8fafc",
				Accent:     "#3b82f6",
			},
			Description: "Discover our curated collection of premium products designed for the modern lifestyle",
			OwnerID:     0,
			Logo:        "default.jpg",
			Banner:      "default_banner.jpg",
			Products: []Product{
				{ID: 1, Name: "Minimalist Phone Case", Description: "Ultra-thin protection with premium materials and clean design.", Price: 29.99, Image: ""},
				{ID: 2, Name: "Smart Watch Series", Description: "Advanced health tracking in a beautifully crafted design.", Price: 199.99, Image: ""},
				{ID: 3, Name: "Wireless Headphones", Description: "Premium sound quality with noise-canceling technology.", Price: 149.99, Image: ""},
				{ID: 4, Name: "Laptop Stand", Description: "Ergonomic aluminum stand for better productivity and posture.", Price: 79.99, Image: ""},
			},
		}
	} else if path == "luxury_template" {
		sampleStore = Store{
			ID:       0,
			Name:     "LUXURIA",
			Template: path,
			ColorScheme: Color{
				Primary:    "#1e3a5f",
				Secondary:  "#36a4b8",
				Background: "#f5f5f0",
			},
			Description: "Exceptional craftsmanship meets timeless elegance in our exclusive collection",
			OwnerID:     0,
			Logo:        "default.jpg",
			Banner:      "default_banner.jpg",
			Products: []Product{
				{ID: 1, Name: "Diamond Heritage Ring", Description: "Swiss-crafted timepiece with hand-set diamonds and platinum casing. A masterpiece of horological excellence.", Price: 2199, Image: "diamond_heritage_ring.jpg"},
				{ID: 2, Name: "Platinum Fountain Pen", Description: "Artisan-crafted writing instrument with 18k gold nib and precious stone accents. A symbol of distinction.", Price: 99, Image: "pen.jpg"},
				{ID: 3, Name: "Crystal Decanter Set", Description: "Hand-blown crystal with gold leaf accents. Perfect for the connoisseur who appreciates the finest spirits.", Price: 200, Image: "decanter_set_blue.jpg"},
			},
		}
	} else if path == "vibrant_template" {
		sampleStore = Store{
			ID:       0,
			Name:     "VibrantShop",
			Template: path,
			ColorScheme: Color{
				Primary:    "#0d9488",
				Secondary:  "#6366f1",
				Tertiary:   "#f59e0b",
				Supporting: "#10b981",
				Highlight:  "#ec4899",
				Accent:     "#8b5cf6",
			},
			Description: "Bold products for creative souls who dare to stand out",
			OwnerID:     0,
			Logo:        "default.jpg",
			Banner:      "default_banner.jpg",
			Products: []Product{
				{ID: 1, Name: "Rainbow Art Set", Description: "Complete creative kit with vibrant colors and premium brushes to unleash your artistic vision.", Price: 45.99, Image: ""},
				{ID: 2, Name: "Holographic Backpack", Description: "Iridescent design that changes colors as you move. Perfect for festival adventures!", Price: 89.99, Image: ""},
				{ID: 3, Name: "LED Galaxy Projector", Description: "Transform any room into a cosmic wonderland with swirling colors and starlight.", Price: 67.99, Image: ""},
				{ID: 4, Name: "Neon Sneakers", Description: "Electric colors that glow in the dark. Step into the future of fashion!", Price: 129.99, Image: ""},
			},
		}
	}
	if err := tmpl.Execute(w, sampleStore); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}

func extractStoreNameFromDomain(hostname string) string {
	if hostname == "" {
		return ""
	}

	parts := strings.Split(hostname, ".")

	if len(parts) >= 3 && parts[len(parts)-2] == "astropify" && parts[len(parts)-1] == "com" {
		return parts[0]
	}

	if len(parts) >= 2 {
		// Try the full domain first
		store, err := GetStoreByCustomDomain(hostname)
		if err == nil && store.ID != 0 {
			return store.Name
		}
		// Return the first part as fallback
		return parts[0]
	}

	return hostname
}
