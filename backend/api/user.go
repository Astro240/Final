package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

func getUser(user_id int) (UserProfile, bool) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return UserProfile{}, false
	}
	defer db.Close()
	var user UserProfile
	query := "SELECT email, first_name || ' ' || last_name, profile_picture from users where id = ?"
	err = db.QueryRow(query, user_id).Scan(&user.Email, &user.Fullname, &user.ProfilePicture)
	if err != nil {
		return UserProfile{}, false
	}
	return user, true
}

func getUserDetails(user_id int) (map[string]interface{}, bool) {
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return nil, false
	}
	defer db.Close()

	var email, firstName, lastName, profilePicture string
	var createdAt time.Time
	query := "SELECT email, first_name, last_name, profile_picture, created_at FROM users WHERE id = ?"
	err = db.QueryRow(query, user_id).Scan(&email, &firstName, &lastName, &profilePicture, &createdAt)
	if err != nil {
		return nil, false
	}

	user := map[string]interface{}{
		"email":           email,
		"first_name":      firstName,
		"last_name":       lastName,
		"fullname":        firstName + " " + lastName,
		"profile_picture": profilePicture,
		"member_since":    createdAt.Format("January 2006"),
	}

	return user, true
}

func ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	user_id, valid := ValidateUser(w, r)
	if !valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userDetails, ok := getUserDetails(user_id)
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := struct {
		User        UserProfile
		FirstName   string
		LastName    string
		MemberSince string
	}{
		User: UserProfile{
			Email:          userDetails["email"].(string),
			Fullname:       userDetails["fullname"].(string),
			ProfilePicture: userDetails["profile_picture"].(string),
		},
		FirstName:   userDetails["first_name"].(string),
		LastName:    userDetails["last_name"].(string),
		MemberSince: userDetails["member_since"].(string),
	}

	tmpl, err := template.ParseFiles("frontend/profile.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id, valid := ValidateUser(w, r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var reqData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer db.Close()

	query := "UPDATE users SET first_name = ?, last_name = ? WHERE id = ?"
	_, err = db.Exec(query, reqData.FirstName, reqData.LastName, user_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update profile"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func UpdateProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user_id, valid := ValidateUser(w, r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File too large"})
		return
	}

	file, handler, err := r.FormFile("profile_picture")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only image files are allowed"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(handler.Filename)
	uuidV4, _ := uuid.NewV4()
	filename := fmt.Sprintf("%s%s", uuidV4.String(), ext)

	// Create avatars directory if it doesn't exist
	avatarsDir := "avatars"
	os.MkdirAll(avatarsDir, os.ModePerm)

	// Save file
	filepath := filepath.Join(avatarsDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save file"})
		return
	}

	// Update database
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	defer db.Close()

	query := "UPDATE users SET profile_picture = ? WHERE id = ?"
	_, err = db.Exec(query, filename, user_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update profile picture"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Profile picture updated successfully",
		"filename": filename,
	})
}
