package api

import (
	"database/sql"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")

	query := "SELECT password FROM users WHERE email = ?"
	row := db.QueryRow(query, email)

	if row.Err() == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	var storedPassword string
	if err := row.Scan(&storedPassword); err != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	SetCookie(w, "session_token")

	// Return success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Login successful"}`))
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Handle registration logic here
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	age := r.FormValue("age")
	//picture can be empty
	profilePicture := r.FormValue("profile_picture")
	//check if email already exists
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		http.Error(w, `{"error": "Database query error"}`, http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		http.Error(w, `{"error": "Email already exists"}`, http.StatusConflict)
		return
	}
	// Validate email and password
	if msg, valid := validateEmail(email); !valid {
		http.Error(w, `{"error": "`+msg+`"}`, http.StatusBadRequest)
		return
	}
	if msg, valid := validatePassword(password); !valid {
		http.Error(w, `{"error": "`+msg+`"}`, http.StatusBadRequest)
		return
	}
	avatarPath := "./avatars/"
	picture := "default.png"
	//check if profilePicture is not empty, then validate and save
	if profilePicture != "" {
		valid, picture := ValidateImage(avatarPath, profilePicture)
		if !valid {
			http.Error(w, `{"error": "`+picture+`"}`, http.StatusInternalServerError)
			return
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Error processing password"}`, http.StatusInternalServerError)
		return
	}
	query := "INSERT INTO users (email, password, first_name, last_name, age, profile_picture) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(query, email, string(hashedPassword), firstName, lastName, age, picture)
	if err != nil {
		http.Error(w, `{"error": "Error creating user"}`, http.StatusInternalServerError)
		return
	}
	SetCookie(w, "session_token")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Registration successful"}`))
}
