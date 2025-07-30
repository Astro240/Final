package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/index.html")
	})
	//handle the src, img, and data directories
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("../frontend/src"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("../frontend/img"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../frontend/data"))))

	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		panic(err)
	}
}