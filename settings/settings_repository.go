package settings

import (
	"log"
	"github.com/JkD004/playarena-backend/db"
)

// GetSetting fetches a value by its key
func GetSetting(key string) (string, error) {
	var value string
	query := `SELECT setting_value FROM site_settings WHERE setting_key = ?`
	err := db.DB.QueryRow(query, key).Scan(&value)
	if err != nil {
		log.Println("Error fetching setting:", err)
		return "", err
	}
	return value, nil
}

// UpdateSetting updates a value by its key
func UpdateSetting(key string, value string) error {
	query := `UPDATE site_settings SET setting_value = ? WHERE setting_key = ?`
	_, err := db.DB.Exec(query, value, key)
	if err != nil {
		log.Println("Error updating setting:", err)
		return err
	}
	return nil
}