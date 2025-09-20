package api

import (
	"database/sql"
	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

func SetCookie(w http.ResponseWriter, id int, name string) {
	sessionID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	query := "INSERT INTO sessions (session_token, user_id) VALUES (?, ?)"
	deleteExisting := "DELETE FROM sessions WHERE user_id = ?"
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	_, err = db.Exec(deleteExisting, id)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(query, sessionID.String(), id)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    sessionID.String(),
		Path:     "/", // ensure cookie is available site-wide
		HttpOnly: true,
		Secure:   false,                // true if HTTPS
		SameSite: http.SameSiteLaxMode, // or NoneMode with Secure for cross-site
		MaxAge:   86400,
	})
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
