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
			store.Products, err = GetProductsByStoreID(store.ID)
			if  err != nil {
				HandleError(w, r, 500, "Couldn't Load Products")
				return
			}
			tmpl, err := template.ParseFiles("../frontend/templates/" + store.Template + "_template.html")
			if err != nil {
				HandleError(w, r, 500, "Couldn't Find Template")
				return
			}
			err = tmpl.Execute(w, store)
			if err != nil {
				HandleError(w, r, 500, "Couldn't Load Template")
				return
			}
			return
		}
	}
	http.ServeFile(w, r, "../frontend/index.html")
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
				Primary:    "#1a202c",
				Secondary:  "#4a5568",
				Background: "#fafafa",
				Accent:     "#667eea",
			},
			Description: "Discover our curated collection of premium products designed for the modern lifestyle",
			OwnerID:     0,
			Logo:        "default.jpg",
			Banner:      "default_banner.jpg",
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
		}
	} else if path == "vibrant_template" {
		sampleStore = Store{
			ID:       0,
			Name:     "VibrantShop",
			Template: path,
			ColorScheme: Color{
				Primary:    "#ff6b6b",
				Secondary:  "#4ecdc4",
				Tertiary:   "#45b7d1",
				Supporting: "#96ceb4",
				Highlight:  "#ffeaa7",
				Accent:     "#fd79a8",
			},
			Description: "Bold products for creative souls who dare to stand out",
			OwnerID:     0,
			Logo:        "default.jpg",
			Banner:      "default_banner.jpg",
		}
	}
	if err := tmpl.Execute(w, sampleStore); err != nil {
		HandleError(w, r, http.StatusInternalServerError, "Failed to render template")
		return
	}
}
