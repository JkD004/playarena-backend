// user/user_service.go
package user

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 1. ADD 'Role' TO THE JWT CLAIMS
type Claims struct {
	UserID int64  `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterNewUser function
func RegisterNewUser(user *User) error {
	if user.Phone != "" {
		_, err := FindUserByPhone(user.Phone)
		if err == nil {
			return errors.New("phone number already in use")
		}
		if err != sql.ErrNoRows {
			log.Println("DB error checking phone:", err)
			return errors.New("database error checking phone number")
		}
	}
	if user.Password == "" || user.Password != user.ConfirmPassword {
		return errors.New("passwords do not match or are empty")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.PasswordHash = string(hashedPassword)
	err = CreateUser(user)
	if err != nil {
		return errors.New("failed to create user, email may be taken")
	}
	return nil
}

// 2. UPDATE LoginUser TO RETURN TOKEN AND ROLE
func LoginUser(email, password string) (string, string, error) {
	storedUser, err := FindUserByEmail(email)
	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: storedUser.ID,
		Email:  storedUser.Email,
		Role:   storedUser.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// --- SECURITY FIX: Read from ENV ---
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// --- FIX WAS HERE: Added extra "" to match signature ---
		return "", "", errors.New("server error: JWT_SECRET not set")
	}
	tokenString, err := token.SignedString([]byte(secret))
	// -----------------------------------
	
	if err != nil {
		return "", "", errors.New("could not generate token")
	}

	return tokenString, storedUser.Role, nil
}

func GetUserProfile(userID int64) (*User, error) {
	return FindUserByID(userID)
}

// UpdateUserProfile handles validation and saving
func UpdateUserProfile(userID int64, updates *User) error {
	// 1. Get existing user to ensure they exist
	_, err := FindUserByID(userID)
	if err != nil {
		return err
	}

	// 2. Set the ID on the update struct so repository knows who to update
	updates.ID = userID
	
	// 3. Save changes
	return UpdateUser(updates)
}

func GetUserByEmail(email string) (*User, error) {
	return FindUserByEmail(email)
}

func UpdateAvatar(userID int64, url string) error {
	return UpdateUserAvatar(userID, url)
}

// user/user_service.go

func GetAllUsers() ([]User, error) {
	return FindAllUsers()
}

func RemoveUser(userID int64) error {
	return DeleteUser(userID)
}

func ChangeUserRole(userID int64, newRole string) error {
	return UpdateUserRoleByID(userID, newRole)
}