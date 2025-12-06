package api

import (
	"database/sql"
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
