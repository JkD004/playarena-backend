// venue/venue_service.go
package venue

import (
	"github.com/JkD004/playarena-backend/db"           // <-- Import db
	"github.com/JkD004/playarena-backend/notification" // <-- Import user
	"github.com/JkD004/playarena-backend/user"
	"log"
	"errors"
	// ... other imports
)

// CreateNewVenue is the business logic for creating a venue.
// It takes the venue data and the ID of the owner from the token.
func CreateNewVenue(venue *Venue, ownerID int64) error {
	// Set the OwnerID on the venue struct
	venue.OwnerID = ownerID

	// You can add validation logic here later
	// (e.g., check if name is empty, price is not negative)

	// Call the repository to save to DB
	err := CreateVenue(venue)
	if err != nil {
		return err
	}

	return nil
}

// ... We're keeping the old GetAllVenues for now,
// but you should update it to fetch from the DB later.
var sampleVenues = []Venue{
	// --- THIS IS THE FIX ---
	// Changed 'Location' to 'Address' to match the struct in venue_model.go
	{ID: 1, Name: "Cantonment Turf", Address: "Belagavi", SportCategory: "Football"},
	// ---------------------
}

// GetAllVenues now fetches ONLY approved venues from the database.
func GetAllVenues() ([]Venue, error) {
	// Call the new repository function
	venues, err := FindApprovedVenues()
	if err != nil {
		return nil, err // Return the error if fetching fails
	}
	return venues, nil
}

// GetVenuesByStatus is the service-layer function to get venues
func GetVenuesByStatus(status string) ([]Venue, error) {
	// You could add validation here, e.g., check if status is a valid value
	return FindVenuesByStatus(status)
}

// UpdateVenueStatus is the service-layer function to update a status
// It now manages a transaction to also upgrade the user's role
func UpdateVenueStatus(venueID int64, newStatus string) error {
	// 1. Start a new database transaction
	tx, err := db.DB.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	// 2. Defer a rollback in case anything goes wrong
	defer tx.Rollback()

	// 3. Get the Owner's ID from the venue
	ownerID, err := GetVenueOwner(tx, venueID)
	if err != nil {
		return err
	}

	// 4. Update the venue status
	err = UpdateVenueStatusInDB(tx, venueID, newStatus)
	if err != nil {
		return err
	}

	// 5. Check if we should upgrade the user's role
	if newStatus == "approved" {
		log.Printf("Venue %d approved. Upgrading user %d to 'owner'", venueID, ownerID)
		err = user.UpdateUserRole(tx, ownerID, "owner")
		if err != nil {
			return err
		}
	}

	if newStatus == "approved" {
		log.Printf("Venue %d approved...", venueID)
		err = user.UpdateUserRole(tx, ownerID, "owner")
		if err != nil {
			return err
		}

		// --- NEW: SEND NOTIFICATION ---
		// Note: Ideally this should be inside the transaction or async,
		// but calling it directly is fine for now.
		_ = notification.CreateNotification(ownerID, "Your venue has been APPROVED! You are now an Owner.", "success")
	} else if newStatus == "rejected" {
		_ = notification.CreateNotification(ownerID, "Your venue listing was rejected by the admin.", "error")
	}

	// 6. If all queries were successful, commit the transaction
	return tx.Commit()
}

// GetVenueByID is the service-layer function to get a single venue
func GetVenueByID(venueID int64) (*Venue, error) {
	return FindApprovedVenueByID(venueID)
}

func GetVenuePhotos(venueID int64) ([]VenuePhoto, error) {
	return GetPhotosByVenueID(venueID)
}

func DeleteVenuePhoto(photoID int64) error {
	// TODO: Add logic to confirm the user (from token) owns this photo
	return DeletePhoto(photoID)
}

// GetVenuesForOwner is the service-layer function
func GetVenuesForOwner(ownerID int64) ([]Venue, error) {
	return FindVenuesByOwnerID(ownerID)
}

// venue/venue_service.go
// ... (keep existing functions)

func AddReview(venueID, userID int64, rating int, comment string) error {
	// TODO: Check if user has actually booked this venue before reviewing?
	review := &Review{
		VenueID: venueID,
		UserID:  userID,
		Rating:  rating,
		Comment: comment,
	}
	return CreateReview(review)
}

// venue/venue_service.go

// venue/venue_service.go

func ModifyVenue(venueID int64, venueData *Venue) error {
	// TODO: Add validation (e.g. ensure price is positive)
	venueData.ID = venueID
	return UpdateVenueDetails(venueData)
}

func GetVenueReviews(venueID int64) ([]Review, error) {
	return GetReviewsByVenueID(venueID)
}

// venue/venue_service.go

func VerifyVenueOwnership(venueID int64, userID int64) error {
	isOwner, err := IsVenueOwner(venueID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("you do not have permission to modify this venue")
	}
	return nil
}
// venue/venue_service.go

// GetVenueIdFromPhoto service wrapper
func GetVenueIdFromPhoto(photoID int64) (int64, error) {
	return GetVenueIDByPhotoID(photoID)
}

func ReplyToReview(reviewID int64, reply string) error {
	// TODO: Check if the logged-in user actually owns the venue this review belongs to
	return AddReviewReply(reviewID, reply)
}