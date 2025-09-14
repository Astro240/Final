package api

import (
	"github.com/gofrs/uuid"
	"net/http"
)

func SetCookie(w http.ResponseWriter, name string) {
	sessionID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	//implementing the cookies that cant be accessed by js or the user
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    sessionID.String(),
		HttpOnly: true,
		Secure:   false,
	})
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
