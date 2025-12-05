// venue/venue_repository.go
package venue

import (
	"database/sql"
	"errors"
	"github.com/JkD004/playarena-backend/db"
	"log"
	"time"
)

// CreateVenue inserts a new venue into the database
func CreateVenue(venue *Venue) error {
	query := `
		INSERT INTO venues (owner_id, name, sport_category, description, address, price_per_hour, opening_time, closing_time, lunch_start_time, lunch_end_time, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending')
	`
	// Handle nullable lunch times
	var lunchStart, lunchEnd sql.NullString
	if venue.LunchStart != "" {
		lunchStart = sql.NullString{String: venue.LunchStart, Valid: true}
	}
	if venue.LunchEnd != "" {
		lunchEnd = sql.NullString{String: venue.LunchEnd, Valid: true}
	}

	result, err := db.DB.Exec(query,
		venue.OwnerID, venue.Name, venue.SportCategory, venue.Description, venue.Address, venue.PricePerHour,
		venue.OpeningTime, venue.ClosingTime, lunchStart, lunchEnd,
	)

	if err != nil {
		log.Println("Error inserting venue:", err)
		return err
	}

	id, _ := result.LastInsertId()
	venue.ID = id
	venue.Status = "pending"
	return nil
}

// GetVenueOwner finds the owner_id of a venue
func GetVenueOwner(tx *sql.Tx, venueID int64) (int64, error) {
	var ownerID int64
	query := `SELECT owner_id FROM venues WHERE id = ?`
	err := tx.QueryRow(query, venueID).Scan(&ownerID)
	if err != nil {
		log.Println("Error finding venue owner:", err)
		return 0, err
	}
	return ownerID, nil
}

// FindVenuesByStatus fetches all venues with a specific status
func FindVenuesByStatus(status string) ([]Venue, error) {
	query := `
		SELECT id, owner_id, status, name, sport_category, description, address, price_per_hour,
		       opening_time, closing_time, lunch_start_time, lunch_end_time, created_at
		FROM venues WHERE status = ?
	`
	rows, err := db.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	venues := make([]Venue, 0)
	for rows.Next() {
		v, err := scanVenue(rows)
		if err == nil {
			venues = append(venues, *v)
		}
	}
	return venues, nil
}

// UpdateVenueStatusInDB updates the status of a specific venue by its ID
// UpdateVenueStatusInDB updates the status using a transaction
func UpdateVenueStatusInDB(tx *sql.Tx, venueID int64, newStatus string) error {
	query := "UPDATE venues SET status = ? WHERE id = ?"

	_, err := tx.Exec(query, newStatus, venueID)
	if err != nil {
		log.Println("Error updating venue status:", err)
		return err
	}
	return nil
}

func FindApprovedVenueByID(venueID int64) (*Venue, error) {
	query := `
		SELECT id, owner_id, status, name, sport_category, description, address, price_per_hour,
		       opening_time, closing_time, lunch_start_time, lunch_end_time, created_at
		FROM venues 
		WHERE id = ? AND status = 'approved'
	`
	rows, err := db.DB.Query(query, venueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanVenue(rows)
	}
	return nil, sql.ErrNoRows
}

func FindApprovedVenues() ([]Venue, error) {
	return FindVenuesByStatus("approved")
}

// GetPhotosByVenueID fetches all photos for a specific venue
func GetPhotosByVenueID(venueID int64) ([]VenuePhoto, error) {
	query := `SELECT id, venue_id, image_url, created_at FROM venue_photos WHERE venue_id = ?`

	rows, err := db.DB.Query(query, venueID)
	if err != nil {
		log.Println("Error fetching venue photos:", err)
		return nil, err
	}
	defer rows.Close()

	var photos []VenuePhoto
	for rows.Next() {
		var photo VenuePhoto
		if err := rows.Scan(&photo.ID, &photo.VenueID, &photo.ImageURL, &photo.CreatedAt); err != nil {
			log.Println("Error scanning venue photo:", err)
			continue
		}
		photos = append(photos, photo)
	}

	if photos == nil {
		photos = make([]VenuePhoto, 0)
	}
	return photos, nil
}

// DeletePhoto deletes a photo by its ID
// TODO: Add check to ensure user owns this photo/venue
func DeletePhoto(photoID int64) error {
	query := `DELETE FROM venue_photos WHERE id = ?`
	_, err := db.DB.Exec(query, photoID)
	if err != nil {
		log.Println("Error deleting photo:", err)
		return errors.New("failed to delete photo")
	}
	return nil
}

// FindVenuesByOwnerID fetches all venues (any status) for a specific owner
func FindVenuesByOwnerID(ownerID int64) ([]Venue, error) {
	query := `
		SELECT id, owner_id, status, name, sport_category, description, address, price_per_hour,
		       opening_time, closing_time, lunch_start_time, lunch_end_time, created_at
		FROM venues WHERE owner_id = ?
	`
	rows, err := db.DB.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	venues := make([]Venue, 0)
	for rows.Next() {
		v, err := scanVenue(rows)
		if err == nil {
			venues = append(venues, *v)
		}
	}
	return venues, nil
}

// venue/venue_repository.go
// ... (keep existing functions)

// CreateReview adds a new review
func CreateReview(review *Review) error {
	query := `INSERT INTO reviews (venue_id, user_id, rating, comment) VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, review.VenueID, review.UserID, review.Rating, review.Comment)
	if err != nil {
		log.Println("Error inserting review:", err)
		return err
	}
	return nil
}

// AddReviewReply updates a review with an owner's reply
func AddReviewReply(reviewID int64, reply string) error {
	query := `UPDATE reviews SET reply = ?, replied_at = ? WHERE id = ?`
	_, err := db.DB.Exec(query, reply, time.Now(), reviewID)
	if err != nil {
		log.Println("Error adding review reply:", err)
		return err
	}
	return nil
}

// GetReviewsByVenueID (Update this existing function to scan the new columns!)
func GetReviewsByVenueID(venueID int64) ([]Review, error) {
	// Updated query to select reply fields
	query := `
		SELECT r.id, r.venue_id, r.user_id, u.first_name, u.last_name, 
		       r.rating, r.comment, r.created_at, 
		       COALESCE(r.reply, ''), r.replied_at
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.venue_id = ?
		ORDER BY r.created_at DESC
	`
	rows, err := db.DB.Query(query, venueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		var repliedAt sql.NullTime // Handle nullable time

		// Updated Scan
		if err := rows.Scan(
			&r.ID, &r.VenueID, &r.UserID, &r.UserFirst, &r.UserLast,
			&r.Rating, &r.Comment, &r.CreatedAt,
			&r.Reply, &repliedAt,
		); err != nil {
			continue
		}

		if repliedAt.Valid {
			r.RepliedAt = repliedAt.Time
		}
		reviews = append(reviews, r)
	}
	// ... (return)
	if reviews == nil {
		reviews = make([]Review, 0)
	}
	return reviews, nil
}

// UpdateVenueDetails updates the text fields of a venue
func UpdateVenueDetails(venue *Venue) error {
	query := `
		UPDATE venues 
		SET name = ?, sport_category = ?, description = ?, address = ?, price_per_hour = ?,
		    opening_time = ?, closing_time = ?, lunch_start_time = ?, lunch_end_time = ?
		WHERE id = ?
	`

	var lunchStart, lunchEnd sql.NullString
	if venue.LunchStart != "" {
		lunchStart = sql.NullString{String: venue.LunchStart, Valid: true}
	}
	if venue.LunchEnd != "" {
		lunchEnd = sql.NullString{String: venue.LunchEnd, Valid: true}
	}

	_, err := db.DB.Exec(query,
		venue.Name, venue.SportCategory, venue.Description, venue.Address, venue.PricePerHour,
		venue.OpeningTime, venue.ClosingTime, lunchStart, lunchEnd,
		venue.ID,
	)
	if err != nil {
		log.Println("Error updating venue details:", err)
		return err
	}
	return nil
}

func scanVenue(rows *sql.Rows) (*Venue, error) {
	var v Venue
	var desc, addr, lStart, lEnd sql.NullString
	var price sql.NullFloat64
	var created sql.NullTime

	err := rows.Scan(
		&v.ID, &v.OwnerID, &v.Status, &v.Name, &v.SportCategory,
		&desc, &addr, &price,
		&v.OpeningTime, &v.ClosingTime, &lStart, &lEnd,
		&created,
	)
	if err != nil {
		return nil, err
	}

	v.Description = desc.String
	v.Address = addr.String
	v.PricePerHour = price.Float64
	v.LunchStart = lStart.String
	v.LunchEnd = lEnd.String
	if created.Valid {
		v.CreatedAt = created.Time
	}

	return &v, nil
}

// IsVenueOwner checks if a specific user owns a specific venue
func IsVenueOwner(venueID int64, ownerID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM venues WHERE id = ? AND owner_id = ?`
	err := db.DB.QueryRow(query, venueID, ownerID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// venue/venue_repository.go

// GetVenueIDByPhotoID finds the venue associated with a photo
func GetVenueIDByPhotoID(photoID int64) (int64, error) {
	var venueID int64
	query := `SELECT venue_id FROM venue_photos WHERE id = ?`
	err := db.DB.QueryRow(query, photoID).Scan(&venueID)
	if err != nil {
		return 0, err
	}
	return venueID, nil
}
