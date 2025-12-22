package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
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

// SetStoreCookie sets a cookie for a specific store path
func SetStoreCookie(w http.ResponseWriter, storeID int, storeName string) {
	sessionID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Store the session token for this store
	query := "INSERT INTO sessions (session_token, user_id) VALUES (?, ?)"
	_, err = db.Exec(query, sessionID.String(), storeID)
	if err != nil {
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Set cookie with store-specific path
	storePath := "/" + strings.ToLower(storeName) + ".com"
	http.SetCookie(w, &http.Cookie{
		Name:     "store_session_" + strings.ToLower(storeName),
		Value:    sessionID.String(),
		Path:     storePath,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
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

// GetStoreCookie retrieves a cookie for a specific store
func GetStoreCookie(r *http.Request, storeName string) (string, error) {
	cookieName := "store_session_" + strings.ToLower(storeName)
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func RemoveCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/", // ensure cookie is available site-wide
		HttpOnly: true,
		Secure:   false,                // true if HTTPS
		SameSite: http.SameSiteLaxMode, // or NoneMode with Secure for cross-site
		MaxAge:   -1,
	})
}

// RemoveStoreCookie removes a cookie for a specific store
func RemoveStoreCookie(w http.ResponseWriter, storeName string) {
	cookieName := "store_session_" + strings.ToLower(storeName)
	storePath := "/" + strings.ToLower(storeName) + ".com"
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     storePath,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
