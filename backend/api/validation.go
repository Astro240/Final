package api

import (
	"regexp"
	"strings"

	"encoding/base64"
	"github.com/gofrs/uuid"
	"os"
)

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
