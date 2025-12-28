package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session token from cookie
	sessionToken, err := GetCookie(r, "session_token")
	if err == nil && sessionToken != "" {
		// Delete the session from database
		db, err := sql.Open("sqlite3", DATABASEPATH)
		if err == nil {
			defer db.Close()
			db.Exec("DELETE FROM sessions WHERE session_token = ?", sessionToken)
		}
	}

	RemoveCookie(w, r, "session_token")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

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
	query := "SELECT id,password FROM users WHERE email = ? AND store_id IS NULL"
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
	firstName := strings.TrimSpace(r.FormValue("first_name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	age := r.FormValue("age")
	ageInt, err := strconv.Atoi(age)
	if err != nil {
		http.Error(w, `{"error": "Age must be a valid number"}`, http.StatusBadRequest)
		return
	}
	if ageInt < 13 {
		http.Error(w, `{"error": "Users must be 13 and above"}`, http.StatusBadRequest)
		return
	}
	if ageInt > 120 {
		http.Error(w, `{"error": "Users must be less than 120"}`, http.StatusBadRequest)
		return
	}
	honeypot := r.FormValue("honeypot")
	if honeypot != "" {
		http.Error(w, `{"error": "Bot detected"}`, http.StatusBadRequest)
		return
	}
	if firstName == "" {
		http.Error(w, `{"error": "First Name Required"}`, http.StatusBadRequest)
		return
	}
	if lastName == "" {
		http.Error(w, `{"error": "Last Name Required"}`, http.StatusBadRequest)
		return
	}
	//picture can be empty
	profilePicture := r.FormValue("profile_picture")
	//check if email already exists
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? and store_id IS NULL", email).Scan(&exists)
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

	avatarPath := AvatarsPath
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

	var userID int
	var storedPassword string
	err = row.Scan(&userID, &storedPassword)

	// If no user found for this store, try to find store owner
	if err == sql.ErrNoRows {
		queryStoreOwner := `SELECT id, password FROM users WHERE email = ? AND store_id IS NULL`
		new_row := db.QueryRow(queryStoreOwner, email)

		if err := new_row.Scan(&userID, &storedPassword); err != nil {
			http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
			http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
			return
		}

		// Successful login for store owner
		SetCookie(w, userID, "session_token")
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		http.Error(w, `{"error": "Invalid email, password, or store ID"}`, http.StatusUnauthorized)
		return
	}

	// Generate and send 2FA code
	token := GenerateEmailCode()
	query = "INSERT INTO verification_codes (user_id, code, expires_at) VALUES (?, ?, datetime('now', '+5 minutes'))"
	_, err = db.Exec(query, userID, token)
	if err != nil {
		http.Error(w, `{"error": "Failed to store verification code"}`, http.StatusInternalServerError)
		return
	}

	err = SendEmail(email, "Verification Code: "+token, "Your store login verification code is: "+token)
	if err != nil {
		http.Error(w, `{"error": "Failed to send verification email"}`, http.StatusInternalServerError)
		return
	}

	// Return success with user_id for 2FA verification
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Verification code sent", "user_id": "` + fmt.Sprint(userID) + `", "email": "` + email + `", "store_id": "` + storeID + `"}`))
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
	email := strings.ToLower(r.FormValue("signupEmail"))
	password := r.FormValue("signupPassword")
	name := r.FormValue("signupName")
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
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? and store_id IS NULL", email).Scan(&exists)
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
	result, err := db.Exec(query, email, string(hashedPassword), name, 2, storeID)
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

	// Generate and send 2FA code
	token := GenerateEmailCode()
	query = "INSERT INTO verification_codes (user_id, code, expires_at) VALUES (?, ?, datetime('now', '+5 minutes'))"
	_, err = db.Exec(query, userID, token)
	if err != nil {
		http.Error(w, `{"error": "Failed to store verification code"}`, http.StatusInternalServerError)
		return
	}

	err = SendEmail(email, "Verification Code: "+token, "Welcome! Your verification code is: "+token)
	if err != nil {
		http.Error(w, `{"error": "Failed to send verification email"}`, http.StatusInternalServerError)
		return
	}

	// Return success with user_id for 2FA verification
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Verification code sent", "user_id": "` + fmt.Sprint(userID) + `", "email": "` + email + `", "store_id": "` + storeID + `"}`))
}

func StoreVerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
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

	userID := r.FormValue("user_id")
	code := r.FormValue("code")
	storeID := r.FormValue("store_id")

	if userID == "" || code == "" || storeID == "" {
		http.Error(w, `{"error": "Missing required fields"}`, http.StatusBadRequest)
		return
	}

	// Verify the code
	var storedCode string
	var expiresAt string
	query := "SELECT code, expires_at FROM verification_codes WHERE user_id = ? ORDER BY created_at DESC LIMIT 1"
	err = db.QueryRow(query, userID).Scan(&storedCode, &expiresAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid or expired verification code"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	if storedCode != code {
		http.Error(w, `{"error": "Invalid verification code"}`, http.StatusUnauthorized)
		return
	}

	// Check if code is expired
	expiry, err := sql.Open("sqlite3", DATABASEPATH)
	if err == nil {
		defer expiry.Close()
		var isExpired bool
		expiry.QueryRow("SELECT datetime('now') > ?", expiresAt).Scan(&isExpired)
		if isExpired {
			http.Error(w, `{"error": "Verification code has expired"}`, http.StatusUnauthorized)
			return
		}
	}

	// Delete used code
	db.Exec("DELETE FROM verification_codes WHERE user_id = ? AND code = ?", userID, code)

	// Set cookies for successful verification
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusInternalServerError)
		return
	}

	customerTokenName := "customer_token_" + storeID
	SetCookie(w, userIDInt, customerTokenName)

	// Set store-specific cookie
	storeIDInt, err := strconv.Atoi(storeID)
	if err != nil {
		http.Error(w, `{"error": "Can't find Store ID"}`, http.StatusInternalServerError)
		return
	}
	store, err := GetStoreByID(storeIDInt)
	if err == nil && store.ID != 0 {
		SetStoreCookie(w, int(store.OwnerID), store.Name)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Verification successful"}`))
}
