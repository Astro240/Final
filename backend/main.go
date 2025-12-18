package main

import (
	"finalProj/api"
	"net/http"
)

func main() {
	api.LoadEnv()
	api.CreateDatabase()
	http.HandleFunc("/", api.HomePage)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/login.html")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/register.html")
	})
	http.HandleFunc("/logout", api.LogoutHandler)
	http.HandleFunc("/store", api.StorePage)
	http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/template.html")
	})
	http.HandleFunc("/templates/preview/", api.SampleStoreView)
	http.HandleFunc("/create-store", func(w http.ResponseWriter, r *http.Request) {
		if _, validUser := api.ValidateUser(w, r); !validUser {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "./frontend/create_store.html")
	})
	http.HandleFunc("/verify_2fa", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/verify_2fa.html")
	})
	http.HandleFunc("/dashboard", api.AdminDashboard)
	//handle the src, img, and data directories
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))
	http.Handle("/logos/", http.StripPrefix("/logos/", http.FileServer(http.Dir("./store_images/logos"))))
	http.Handle("/banners/", http.StripPrefix("/banners/", http.FileServer(http.Dir("./store_images/banners"))))
	http.Handle("/products/", http.StripPrefix("/products/", http.FileServer(http.Dir("./store_images/products"))))

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./frontend/src"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./frontend/img"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./frontend/data"))))

	// Favicon handler
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/favicon.ico")
	})

	//======= APIs for the project =======//

	// Registration-Login Hanlders //
	http.HandleFunc("/api/login", api.LoginHandler)
	http.HandleFunc("/api/store_login", api.StoreLoginHandler)
	http.HandleFunc("/api/register", api.RegisterHandler)
	http.HandleFunc("/api/store_register", api.StoreRegisterHandler)
	http.HandleFunc("/api/verify_2fa", api.TwoFactorAuth)
	http.HandleFunc("/api/resend_2fa", api.Resend2FAHandler)
	// Store-Product Handlers //
	http.HandleFunc("/api/create_store", api.CreateStoreHandler)
	http.HandleFunc("/api/create_product", api.CreateProductAPI)
	http.HandleFunc("/api/favorite_product", api.FavoriteProduct)
	http.HandleFunc("/api/unfavorite_product", api.UnfavoriteProduct)
	http.HandleFunc("/api/products", api.GetProductsAPI)
	http.HandleFunc("/api/products/update", api.UpdateProductAPI)
	http.HandleFunc("/api/products/delete", api.DeleteProductAPI)
	// Cart and Checkout routes //
	http.HandleFunc("/api/add-to-cart", api.AddToCart)
	http.HandleFunc("/api/get-cart", api.GetCartTotal)
	http.HandleFunc("/api/update-cart", api.UpdateCartItem)
	http.HandleFunc("/api/remove-from-cart", api.RemoveFromCart)
	http.HandleFunc("/api/create-order", api.CreateOrder)
	// Payment routes
	http.HandleFunc("/api/pending-order", api.GetPendingOrder)
	http.HandleFunc("/api/process-payment", api.ProcessPayment)
	// Store Orders routes
	http.HandleFunc("/api/orders", api.GetStoreOrders)
	http.HandleFunc("/api/orders/update-status", api.UpdateOrderStatus)
	http.HandleFunc("/api/orders/", api.GetOrderProducts)
	http.HandleFunc("/api/my-orders", api.GetOrders)
	http.HandleFunc("/api/customer/orders", api.GetCustomerOrders)
	// WebSocket routes
	http.HandleFunc("/ws/dashboard", api.DashboardWebSocketHandler)

	// Admin API routes
	http.HandleFunc("/api/admin/stats", api.AdminStats)
	http.HandleFunc("/api/admin/stores", api.AdminStores)
	http.HandleFunc("/api/admin/users", api.AdminUsers)
	http.HandleFunc("/api/admin/orders", api.AdminOrders)
	http.HandleFunc("/api/admin/products", api.AdminProducts)

	if err := http.ListenAndServe("0.0.0.0:80", nil); err != nil {
		panic(err)
	}
}
