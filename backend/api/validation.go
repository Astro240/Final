package api

import (
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"
	"fmt"
	"database/sql"
	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func ValidateUser(w http.ResponseWriter, r *http.Request) (int,bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}
	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return 0, false
	}
	defer db.Close()
	var userID int
	query := "SELECT user_id FROM sessions WHERE session_token = ?"
	row := db.QueryRow(query, cookie.Value)
	if row.Err() == sql.ErrNoRows {
		return 0, false
	}
	if err := row.Scan(&userID); err != nil {
		return 0, false
	}
	return userID, true
}

func validateEmail(email string) (string, bool) {
	if len(email) < 5 || len(email) > 50 {
		return "Email length must be between 5 and 50 characters", false
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return "Invalid email format", false
	}
	return "", true
}

func validatePassword(password string) (string, bool) {
	if len(password) < 8 || len(password) > 50 {
		return "Password length must be between 8 and 50 characters", false
	}

	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	specialChars := `@$!%*?&`

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}
	if !(hasLower && hasUpper && hasDigit && hasSpecial) {
		return "Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character", false
	}
	return "", true
}

func ValidateImage(path string, avatar string) (bool, string) {
	photoType := ".png"
	err := ""
	if !strings.HasPrefix(avatar, "data:image/png;base64,") &&
		!strings.HasPrefix(avatar, "data:image/jpeg;base64,") &&
		!strings.HasPrefix(avatar, "data:image/webp;base64,") &&
		!strings.HasPrefix(avatar, "data:image/bmp;base64,") &&
		!strings.HasPrefix(avatar, "data:image/vnd.microsoft.icon;base64,") && // ICO format
		!strings.HasPrefix(avatar, "data:image/gif;base64,") { // GIF format
		err = "Image should be either 'png', 'jpg', 'webp', 'bmp', 'ico', or 'gif'"
		return false, err
	}
	if strings.HasPrefix(avatar, "data:image/gif;base64,") {
		photoType = ".gif"
	}
	data := strings.Split(avatar, ",")[1] // Get the Base64 part
	fileData, errorBase := base64.StdEncoding.DecodeString(data)
	if errorBase != nil {
		err = "Error decoding Base64"
		return false, err
	}
	if len(fileData) >= 1*1024*1024 { // 1 MB
		err = "Image size should not exceed 1 MB"
		return false, err
	}
	// Save the file (you can customize the file name and path)
	uuid, erroruuid := uuid.NewV4()
	if erroruuid != nil {
		err = "Unable to Generate UUID"
		return false, err
	}
	errorOs := os.WriteFile(path+uuid.String()+photoType, fileData, 0644)
	if errorOs != nil {
		err = "Error writing file"
		return false, err
	}
	avatar = uuid.String() + photoType
	return true, avatar
}

func ValidateStoreName(name string) error {
	if strings.Contains(name, " ") || strings.Contains(name, ",") {
		return fmt.Errorf("Store name cannot contain spaces or commas")
	}
	
	invalidChars := regexp.MustCompile(`[!@#$%^&*()_+=\[\]{}|\\:;"'<>,?/\u200B]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("Store name contains invalid characters")
	}

	htmlTags := regexp.MustCompile(`<[^>]*>`)
	if htmlTags.MatchString(name) {
		return fmt.Errorf("Store name cannot contain HTML tags")
	}

	if strings.TrimSpace(name) != name {
		return fmt.Errorf("Store name cannot have leading or trailing whitespace")
	}

	return nil
}