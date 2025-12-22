package api

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func ValidateUser(w http.ResponseWriter, r *http.Request) (int, bool) {
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

func ValidateCustomer(w http.ResponseWriter, r *http.Request) (int, bool) {
	var storeName string

	// Try extracting from custom header first (for API calls)
	headerValue := r.Header.Get("X-Store-Name")
	if headerValue != "" {
		// Parse header value which could be a full URL or just path
		// e.g., "http://astropify.com/LUXURY.com" or "/LUXURY.com"
		if strings.Contains(headerValue, "://") {
			// It's a full URL, extract the path
			parts := strings.Split(headerValue, "://")
			if len(parts) > 1 {
				pathPart := parts[1]
				// Remove domain part (everything before first /)
				if idx := strings.Index(pathPart, "/"); idx != -1 {
					pathPart = pathPart[idx:]
				} else {
					pathPart = ""
				}
				headerValue = pathPart
			}
		}

		// Now extract store name from path
		if headerValue != "" {
			pathParts := strings.Split(strings.Trim(headerValue, "/"), "/")
			if len(pathParts) > 0 {
				candidate := pathParts[0]
				// Remove .com or other extensions
				if strings.Contains(candidate, ".") {
					dotParts := strings.Split(candidate, ".")
					if len(dotParts) > 0 {
						storeName = dotParts[0]
					}
				} else {
					storeName = candidate
				}
			}
		}
	}

	// Try query/form parameters
	if storeName == "" {
		storeName = r.URL.Query().Get("store_name")
	}
	if storeName == "" {
		storeName = r.FormValue("store_name")
	}

	// Try extracting from hostname (for direct domain access)
	if storeName == "" {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		if strings.Contains(host, ".") && !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") && !strings.HasPrefix(host, "astropify") {
			parts := strings.Split(host, ".")
			if len(parts) > 0 {
				storeName = parts[0]
			}
		}
	}

	// If not found, extract from URL path (skip "api" and common route prefixes)
	if storeName == "" {
		path := r.URL.Path
		pathParts := strings.Split(strings.Trim(path, "/"), "/")

		if len(pathParts) > 0 {
			firstPart := pathParts[0]
			// Remove .com or other extensions if present
			if strings.Contains(firstPart, ".") {
				parts := strings.Split(firstPart, ".")
				if len(parts) > 0 {
					storeName = parts[0]
				}
			} else {
				storeName = firstPart
			}
		}
	}

	if storeName == "" {
		return 0, false
	}

	db, err := sql.Open("sqlite3", DATABASEPATH)
	if err != nil {
		return 0, false
	}
	defer db.Close()

	// Get store ID from store name
	var storeID int
	err = db.QueryRow("SELECT id FROM stores WHERE name = ?", storeName).Scan(&storeID)
	if err != nil {
		return 0, false
	}
	// Use store-specific customer token
	storeIDStr := fmt.Sprintf("%d", storeID)
	return ValidateStoreCustomer(w, r, storeIDStr)
}

func ValidateStoreCustomer(w http.ResponseWriter, r *http.Request, storeID string) (int, bool) {
	customerTokenName := "customer_token_" + storeID
	cookie, err := r.Cookie(customerTokenName)
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

	invalidChars := regexp.MustCompile(`[!@#$%^&*()_+=\[\]{}|\\:;"'<>,?/\x{200B}]`)
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
