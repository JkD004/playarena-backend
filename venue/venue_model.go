// venue/venue_model.go
package venue

import "time"

type Venue struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	Status        string    `json:"status"`
	Name          string    `json:"name"`
	SportCategory string    `json:"sport_category"`
	Description   string    `json:"description,omitempty"`
	Address       string    `json:"address,omitempty"`
	PricePerHour  float64   `json:"price_per_hour,omitempty"`
	OpeningTime   string    `json:"opening_time"`
	ClosingTime   string    `json:"closing_time"`
	LunchStart    string    `json:"lunch_start_time,omitempty"`
	LunchEnd      string    `json:"lunch_end_time,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	
}
// VenuePhoto defines the data structure for a photo
type VenuePhoto struct {
	ID        int64     `json:"id"`
	VenueID   int64     `json:"venue_id"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}


// venue/venue_model.go

// VenueAdminList is for the main list view
type VenueAdminList struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	SportCategory string `json:"sport_category"`
	Status        string `json:"status"`
	OwnerName     string `json:"owner_name"`
}

// VenueAdminDetail contains everything about a venue
type VenueAdminDetail struct {
	VenueID       int64   `json:"venue_id"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	SportCategory string  `json:"sport_category"`
	PricePerHour  float64 `json:"price_per_hour"`
	Status        string  `json:"status"`
	// Owner Details
	OwnerID       int64   `json:"owner_id"`
	OwnerName     string  `json:"owner_name"`
	OwnerEmail    string  `json:"owner_email"`
	OwnerPhone    string  `json:"owner_phone"`
}


type Review struct {
	ID        int64     `json:"id"`
	VenueID   int64     `json:"venue_id"`
	UserID    int64     `json:"user_id"`
	UserFirst string    `json:"user_first_name"` // To show who wrote it
	UserLast  string    `json:"user_last_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	Reply     string    `json:"reply,omitempty"`      // <-- NEW
    RepliedAt time.Time `json:"replied_at,omitempty"`
}