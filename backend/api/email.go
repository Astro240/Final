package api

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
)

func GenerateEmailCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	return code
}

func SendEmail(to string, subject string, body string) error {
	gmailKey := os.Getenv("GMAIL_KEY")
	auth := smtp.PlainAuth("", "astropify@gmail.com", gmailKey, "smtp.gmail.com")

	// List of providers to try
	providers := map[string]string{
		"gmail":   "smtp.gmail.com:587",
		"yahoo":   "smtp.mail.yahoo.com:587",
		"outlook": "smtp.office365.com:587",
		"hotmail": "smtp.live.com:587",
	}

	for _, server := range providers {
		err := smtp.SendMail(server, auth, "astropify@gmail.com", []string{to}, []byte("Subject: "+subject+"\n\n"+body))
		if err != nil {
			continue
		}
		return nil // Exit after a successful send
	}
	return fmt.Errorf("failed to send email to %s: all SMTP providers failed", to)
}

func TwoFactorAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.FormValue("user_id")
	code := r.FormValue("code")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	_, _ = db.Exec("DELETE FROM verification_codes WHERE expires_at <= datetime('now')")
	var userIDInt int
	query := "SELECT id,user_id FROM verification_codes WHERE user_id = ? AND code = ? AND expires_at > datetime('now')"
	row := db.QueryRow(query, userID, code)

	var id int
	if err := row.Scan(&id, &userIDInt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "Invalid or expired code"}`, http.StatusUnauthorized)
		} else {
			http.Error(w, `{"error": "Database query error"}`, http.StatusInternalServerError)
		}
		return
	}
	_, _ = db.Exec("DELETE FROM verification_codes WHERE user_id = ? AND (code = ? OR expires_at <= datetime('now'))", userIDInt, code)
	SetCookie(w, userIDInt, "session_token")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Login successful"}`))
}

func Resend2FAHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.FormValue("user_id")

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	var email string
	query := "SELECT u.email,v.code from verification_codes v JOIN users u ON v.user_id = u.id WHERE v.user_id = ? AND v.expires_at > datetime('now')"
	row := db.QueryRow(query, userID)

	var existingCode string
	if err := row.Scan(&email, &existingCode); err == nil {
		err = SendEmail(email, "Your 2FA Code", "Your verification code is: "+existingCode)
		if err != nil {
			http.Error(w, `{"error": "Failed to send email"}`, http.StatusInternalServerError)
			return
		}
		http.Error(w, `{"error": "A valid code has already been sent. Please check your email."}`, http.StatusBadRequest)
		return
	} else {
		token := GenerateEmailCode()

		query = "INSERT INTO verification_codes (user_id, code, expires_at) VALUES (?, ?, datetime('now', '+5 minutes'))"
		_, err = db.Exec(query, userID, token)
		if err != nil {
			http.Error(w, `{"error": "Failed to generate new code"}`, http.StatusInternalServerError)
			return
		}
		err = SendEmail(email, "Your 2FA Code", "Your verification code is: "+token)
		if err != nil {
			http.Error(w, `{"error": "Failed to send email"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Verification code resent successfully","reset_timer": true}`))
	}
}
