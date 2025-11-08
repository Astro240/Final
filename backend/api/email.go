package api

import (
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"database/sql"
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

	err := smtp.SendMail("smtp.gmail.com:587", auth, "astropify@gmail.com", []string{to}, []byte("Subject: "+subject+"\n\n"+body))
	return err
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
