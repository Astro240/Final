package api

import (
	"html/template"
	"net/http"
	"strings"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	stores, err := GetStores()
	if err != nil {
		http.ServeFile(w, r, "../frontend/index.html")
		return
	}
	for _, store := range stores {
		if r.URL.Path == "/"+strings.ToLower(store.Name)+".com" || r.Host == strings.ToLower(store.Name)+".com" {
			//serve the store for the store
			tmpl, err := template.ParseFiles("../frontend/templates/" + store.Template + "_template.html")
    		if err != nil {
    			HandleError(w,r,500,"Couldn't Find Template")
				return
    		}
			tmpl.Execute(w, store)
			return
		}
	}
	http.ServeFile(w, r, "../frontend/index.html")
}