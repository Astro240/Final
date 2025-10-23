package main

import (
	"finalProj/api"
	"net/http"
)

func main() {
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
	http.HandleFunc("/create_store", api.CreateStoreHandler)
	//handle the src, img, and data directories
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./backend/avatars"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("../frontend/src"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("../frontend/img"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../frontend/data"))))

	http.HandleFunc("/api/login", api.LoginHandler)
	http.HandleFunc("/api/register", api.RegisterHandler)
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		panic(err)
	}
}
