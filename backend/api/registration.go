package api

import (
	"database/sql"
	"fmt"
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
	honeypot := r.FormValue("honeypot")
	if honeypot != "" {
		http.Error(w, `{"error": "Bot detected"}`, http.StatusBadRequest)
		return
	}
	query := "SELECT id,password FROM users WHERE email = ?"
	row := db.QueryRow(query, email)

	if row.Err() == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	var userID int
	var storedPassword string
	if err := row.Scan(&userID, &storedPassword); err != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	token := GenerateEmailCode()

	query = "INSERT INTO verification_codes (user_id, code, expires_at) VALUES (?, ?, datetime('now', '+5 minutes'))"
	_, err = db.Exec(query, userID, token)
	if err != nil {
		http.Error(w, `{"error": "Failed to store verification code"}`, http.StatusInternalServerError)
		return
	}
	err = SendEmail(email, "Verification Code To Astropify: "+token, "A login to your account was just made. If this wasn't you, please reset your password immediately.")
	if err != nil {
		http.Error(w, `{"error": "Failed to send email"}`, http.StatusInternalServerError)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Login successful","user_id": "` + fmt.Sprint(userID) + `", "email": "` + email + `"}`))
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
	honeypot := r.FormValue("honeypot")
	if honeypot != "" {
		http.Error(w, `{"error": "Bot detected"}`, http.StatusBadRequest)
		return
	}
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

	token := GenerateEmailCode()
	err = SendEmail(email, "Verification Code To Astropify: "+token, "Welcome to Astropify, Where you can create your store on the fly! Your verification code is: "+token)
	if err != nil {
		http.Error(w, `{"error": "Failed to send email"}`, http.StatusInternalServerError)
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

	query := "INSERT INTO users (email, password, first_name, last_name, age,user_type, profile_picture) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, email, string(hashedPassword), firstName, lastName, age, 2, picture)
	if err != nil {
		http.Error(w, `{"error": "Error creating user"}`, http.StatusInternalServerError)
		return
	}

	// Get the last inserted ID
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, `{"error": "Error retrieving user ID"}`, http.StatusInternalServerError)
		return
	}

	query = "INSERT INTO verification_codes (user_id, code, expires_at) VALUES (?, ?, datetime('now', '+5 minutes'))"
	_, err = db.Exec(query, userID, token)
	if err != nil {
		http.Error(w, `{"error": "Failed to store verification code"}`, http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Registration successful", "user_id": "` + fmt.Sprint(userID) + `", "email": "` + email + `"}`))
}

func StoreLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	email := strings.ToLower(r.FormValue("loginEmail"))
	password := r.FormValue("loginPassword")
	storeID := r.FormValue("store_id")
	honeypot := r.FormValue("honeypot")
	if honeypot != "" {
		http.Error(w, `{"error": "Bot detected"}`, http.StatusBadRequest)
		return
	}
	query := `SELECT id, password FROM users WHERE email = ? AND store_id = ?`
	row := db.QueryRow(query, email, storeID)
	if row.Err() == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
		return
	}

	var userID int
	var storedPassword string
	if err := row.Scan(&userID, &storedPassword); err != nil {
		http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
		return
	}
	// Successful login
	w.WriteHeader(http.StatusOK)
}

func StoreRegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	name := r.FormValue("name")
	storeID := r.FormValue("store_id")
	honeypot := r.FormValue("honeypot")

	if honeypot != "" {
		http.Error(w, `{"error": "Bot detected"}`, http.StatusBadRequest)
		return
	}
	// Check if store exists and get owner_id
	var ownerID int
	err = db.QueryRow("SELECT owner_id FROM stores WHERE id = ?", storeID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Store does not exist"}`, http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Database query error"}`, http.StatusInternalServerError)
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
	// Check if email already exists
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? and store_id = ?", email, storeID).Scan(&exists)
	if err != nil {
		http.Error(w, `{"error": "Database query error"}`, http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		http.Error(w, `{"error": "Email already exists"}`, http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Error processing password"}`, http.StatusInternalServerError)
		return
	}
	query := "INSERT INTO users (email, password, first_name, user_type, store_id) VALUES (?, ?, ?, ?, ?)"
	_, err = db.Exec(query, email, string(hashedPassword), name, 2, storeID)
	if err != nil {
		http.Error(w, `{"error": "Error creating user"}`, http.StatusInternalServerError)
		return
	}
	// Get the last inserted ID
	// userID, err := result.LastInsertId()
	// if err != nil {
	// 	http.Error(w, `{"error": "Error retrieving user ID"}`, http.StatusInternalServerError)
	// 	return
	// }
	// Successful registration
	w.WriteHeader(http.StatusOK)
}
