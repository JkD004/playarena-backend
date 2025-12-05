// user/user_repository.go
package user

import (
	"database/sql" // We need this
	"log"
	//"time"

	"github.com/JkD004/playarena-backend/db"
)

// --- CreateUser (No changes) ---
func CreateUser(user *User) error {
	query := `INSERT INTO users (first_name, last_name, phone, dob, address, email, password_hash)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
			  
	_, err := db.DB.Exec(query, user.FirstName, user.LastName, user.Phone, user.DOB, user.Address, user.Email, user.PasswordHash)
	
	if err != nil {
		log.Println("Error inserting user:", err)
		return err
	}
	return nil
}

// --- FindUserByEmail (No changes) ---
func FindUserByEmail(email string) (*User, error) {
	var user User
	query := "SELECT id, email, password_hash, first_name, last_name, role FROM users WHERE email = ?"
	err := db.DB.QueryRow(query, email).Scan(
		&user.ID, 
		&user.Email, 
		&user.PasswordHash, 
		&user.FirstName, 
		&user.LastName, 
		&user.Role,
	)
	if err != nil {
		log.Println("Error finding user by email:", err)
		return nil, err
	}
	return &user, nil
}

// --- FindUserByPhone (No changes) ---
func FindUserByPhone(phone string) (*User, error) {
	var user User
	query := "SELECT id FROM users WHERE phone = ?"
	err := db.DB.QueryRow(query, phone).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}



// FindUserByID fetches a user's profile (excluding password)
// FindUserByID fetches a user's profile
func FindUserByID(userID int64) (*User, error) {
	var user User
	// Updated Query to include avatar_url (using COALESCE to handle NULLs)
	safeQuery := `
		SELECT id, first_name, last_name, email, 
		COALESCE(phone, ''), COALESCE(dob, ''), COALESCE(address, ''), 
		role, created_at, COALESCE(avatar_url, '') 
		FROM users 
		WHERE id = ?
	`
	// Updated Scan
	err := db.DB.QueryRow(safeQuery, userID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, 
		&user.Phone, &user.DOB, &user.Address, 
		&user.Role, &user.CreatedAt, &user.AvatarURL, // <-- Added AvatarURL
	)
	
	if err != nil {
		log.Println("Error finding user by ID:", err)
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's profile info
func UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET first_name = ?, last_name = ?, phone = ?, dob = ?, address = ? 
		WHERE id = ?
	`
	_, err := db.DB.Exec(query, 
		user.FirstName, user.LastName, user.Phone, 
		user.DOB, user.Address, user.ID,
	)
	if err != nil {
		log.Println("Error updating user:", err)
		return err
	}
	return nil
}

// UpdateUserRole updates a user's role in the database
// It uses a transaction (tx) to ensure data integrity
func UpdateUserRole(tx *sql.Tx, userID int64, newRole string) error {
	query := `UPDATE users SET role = ? WHERE id = ? AND role = 'player'`
	
	// We only update if the user is currently a 'player'
	// This prevents an admin from being demoted if they submit a venue
	result, err := tx.Exec(query, newRole, userID)
	if err != nil {
		log.Println("Error updating user role:", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Updated role for user %d to %s. Rows affected: %d", userID, newRole, rowsAffected)
	
	return nil
}

// user/user_repository.go

// UpdateUserAvatar updates the avatar_url for a user
func UpdateUserAvatar(userID int64, url string) error {
	query := `UPDATE users SET avatar_url = ? WHERE id = ?`
	_, err := db.DB.Exec(query, url, userID)
	if err != nil {
		log.Println("Error updating user avatar:", err)
		return err
	}
	return nil
}

// FindAllUsers fetches every user in the database
func FindAllUsers() ([]User, error) {
	// We exclude password_hash for security
	query := `
		SELECT id, first_name, last_name, email, phone, dob, address, role, created_at, COALESCE(avatar_url, '')
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.FirstName, &u.LastName, &u.Email, 
			&u.Phone, &u.DOB, &u.Address, 
			&u.Role, &u.CreatedAt, &u.AvatarURL,
		); err != nil {
			continue
		}
		users = append(users, u)
	}
	
	if users == nil {
		users = make([]User, 0)
	}
	return users, nil
}

// DeleteUser permanently removes a user
func DeleteUser(userID int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.DB.Exec(query, userID)
	return err
}

// UpdateUserRoleByID directly updates a role (wrapper for existing logic if needed, 
// but we can reuse UpdateUserRole if we have a transaction, or just run a simple query here)
func UpdateUserRoleByID(userID int64, newRole string) error {
	query := `UPDATE users SET role = ? WHERE id = ?`
	_, err := db.DB.Exec(query, newRole, userID)
	return err
}