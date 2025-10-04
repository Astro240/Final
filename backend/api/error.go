package api

import (
	"html/template"
	"net/http"
)

type ErrorResponse struct {
	ErrorNbr int    `json:"error_number"`
	ErrorMsg string `json:"error_message"`
}

func HandleError(w http.ResponseWriter, r *http.Request, errNbr int, err error) {
	tmpl, tmplErr := template.ParseFiles("../frontend/templates/Error.html")
	if tmplErr != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	errorMsg := "Unknown error"
    if err != nil {
        errorMsg = err.Error()
    }
	errorData := ErrorResponse{
		ErrorNbr: errNbr,
		ErrorMsg: errorMsg,
	}

	if execErr := tmpl.Execute(w, errorData); execErr != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
