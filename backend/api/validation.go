package api

import (
	"regexp"
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
	re := regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`)
	if !re.MatchString(password) {
		return "Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character", false
	}
	return "", true
}
