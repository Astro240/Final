package api

import (
	"html/template"
	"net/http"
)

func HandleError(w http.ResponseWriter, r *http.Request, errNbr int, err string) {
	tmpl, tmplErr := template.ParseFiles(FrontendErrorHTML)
	if tmplErr != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	errorData := ErrorResponse{
		ErrorNbr: errNbr,
		ErrorMsg: err,
	}
	w.WriteHeader(errNbr)
	if execErr := tmpl.Execute(w, errorData); execErr != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
