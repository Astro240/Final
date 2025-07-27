package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//serve the index.html file
		http.ServeFile(w, r, "../frontend/index.html")
	})
	//handle all the css and js files in the src
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("../frontend/src"))))
	//handle all the images in the img folder
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("../frontend/img"))))
	//serve the json file for the frontend
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("../frontend/data"))))

	http.ListenAndServe("0.0.0.0:8080", nil)
}