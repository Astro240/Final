package main

import (
	"finalProj/api"
	"net/http"
)

func main() {
	api.LoadEnv()
	api.CreateDatabase()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/index.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/login.html")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/register.html")
	})
	http.HandleFunc("/store", api.StorePage)
	http.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/template.html")
	})
	http.HandleFunc("/templates/preview/", api.SampleStoreView)
	http.HandleFunc("/create-store", func(w http.ResponseWriter, r *http.Request) {
		if _, validUser := api.ValidateUser(w, r); !validUser {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "../frontend/create_store.html")
	})
	http.HandleFunc("/verify_2fa", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/verify_2fa.html")
	})
	//handle the src, img, and data directories
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./backend/avatars"))))
	http.Handle("/logos/", http.StripPrefix("/store_images/logos/", http.FileServer(http.Dir("./backend/store_images/logos"))))
	http.Handle("/banners/", http.StripPrefix("/store_images/banners/", http.FileServer(http.Dir("./backend/store_images/banners"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("../frontend/src"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("../frontend/img"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../frontend/data"))))

	//APIs for the project
	http.HandleFunc("/api/login", api.LoginHandler)
	http.HandleFunc("/api/register", api.RegisterHandler)
	http.HandleFunc("/api/verify_2fa", api.TwoFactorAuth)
	http.HandleFunc("/api/create_store", api.CreateStoreHandler)
	http.HandleFunc("/api/resend_2fa", api.Resend2FAHandler)

	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		panic(err)
	}
}
