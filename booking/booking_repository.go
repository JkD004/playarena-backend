// booking/booking_repository.go
package booking

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/JkD004/playarena-backend/db"
)

// CreateBooking inserts a new booking into the database
func CreateBooking(booking *Booking) error {
	query := `
		INSERT INTO bookings (user_id, venue_id, start_time, end_time, total_price, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := db.DB.Exec(query,
		booking.UserID,
		booking.VenueID,
		booking.StartTime,
		booking.EndTime,
		booking.TotalPrice,
		booking.Status,
	)
	if err != nil {
		log.Println("Error inserting booking:", err)
		return err
	}

	id, _ := result.LastInsertId()
	booking.ID = id
	return nil
}

// IsSlotAvailable checks for overlapping bookings
func IsSlotAvailable(venueID int64, startTime, endTime time.Time) (bool, error) {
	var count int
    // FIX: Only block slots that are 'confirmed' or 'present'.
    // 'canceled' AND 'absent' slots should be ignored (available).
	query := `
		SELECT COUNT(*) FROM bookings
		WHERE venue_id = ?
		AND status IN ('confirmed', 'present') 
		AND start_time < ?
		AND end_time > ?
	`
	
	err := db.DB.QueryRow(query, venueID, endTime, startTime).Scan(&count)
	if err != nil {
		log.Println("Error checking slot availability:", err)
		return false, err
	}
	return count == 0, nil
}

// FindBookingsByUserID fetches all bookings for a specific user
func FindBookingsByUserID(userID int64) ([]Booking, error) {
	// JOIN venues to get Name and Sport
	query := `
		SELECT 
			b.id, b.user_id, b.venue_id, 
			v.name, v.sport_category, 
			b.start_time, b.end_time, b.total_price, b.status, b.created_at
		FROM bookings b
		JOIN venues v ON b.venue_id = v.id
		WHERE b.user_id = ?
		ORDER BY b.start_time DESC
	`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		log.Println("Error querying bookings by user ID:", err)
		return nil, err
	}
	defer rows.Close()

	bookings := make([]Booking, 0)
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.VenueID,
			&booking.VenueName,     
			&booking.SportCategory, 
			&booking.StartTime,
			&booking.EndTime,
			&booking.TotalPrice,
			&booking.Status,
			&booking.CreatedAt,
		); err != nil {
			log.Println("Error scanning booking row:", err)
			continue
		}
		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

// UpdateBookingStatus updates a booking's status
func UpdateBookingStatus(bookingID int64, userID int64, newStatus string) error {
	query := `
		UPDATE bookings 
		SET status = ? 
		WHERE id = ? AND user_id = ?
	`

	result, err := db.DB.Exec(query, newStatus, bookingID, userID)
	if err != nil {
		log.Println("Error updating booking status:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("booking not found or you do not have permission")
	}
	return nil
}

// FindBookingsByVenueID fetches all bookings for a specific venue, including user info
func FindBookingsByVenueID(venueID int64) ([]AdminBookingView, error) {
	query := `
		SELECT 
			b.id, b.venue_id, v.name, v.sport_category, b.user_id, 
			u.first_name, u.last_name, COALESCE(u.phone, 'N/A'),
			b.start_time, b.end_time, b.total_price, b.status
		FROM bookings b
		JOIN venues v ON b.venue_id = v.id
		JOIN users u ON b.user_id = u.id 
		WHERE b.venue_id = ?
		ORDER BY b.start_time DESC
	`
	rows, err := db.DB.Query(query, venueID)
	if err != nil {
		log.Println("Error querying bookings by venue:", err)
		return nil, err
	}
	defer rows.Close()

	var bookings []AdminBookingView
	for rows.Next() {
		var b AdminBookingView
		if err := rows.Scan(
			&b.BookingID,
			&b.VenueID,
			&b.VenueName,
			&b.SportCategory,
			&b.UserID,
			&b.UserFirstName,
			&b.UserLastName,
			&b.UserPhone, 
			&b.StartTime,
			&b.EndTime,
			&b.TotalPrice,
			&b.Status,
		); err != nil {
			log.Println("Error scanning booking row:", err)
			continue
		}
		bookings = append(bookings, b)
	}

	if bookings == nil {
		bookings = make([]AdminBookingView, 0)
	}
	return bookings, nil
}

// FindAllBookings fetches a master list of all bookings
func FindAllBookings() ([]AdminBookingView, error) {
	query := `
		SELECT 
			b.id, b.venue_id, v.name, v.sport_category, b.user_id, 
			u.first_name, u.last_name, COALESCE(u.phone, 'N/A'),
			b.start_time, b.end_time, b.total_price, b.status
		FROM bookings b
		JOIN venues v ON b.venue_id = v.id
		JOIN users u ON b.user_id = u.id 
		ORDER BY b.start_time DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Println("Error querying all bookings:", err)
		return nil, err
	}
	defer rows.Close()

	var bookings []AdminBookingView
	for rows.Next() {
		var b AdminBookingView
		if err := rows.Scan(
			&b.BookingID,
			&b.VenueID,
			&b.VenueName,
			&b.SportCategory,
			&b.UserID,
			&b.UserFirstName,
			&b.UserLastName,
			&b.UserPhone,
			&b.StartTime,
			&b.EndTime,
			&b.TotalPrice,
			&b.Status,
		); err != nil {
			log.Println("Error scanning booking row:", err)
			continue
		}
		bookings = append(bookings, b)
	}

	if bookings == nil {
		bookings = make([]AdminBookingView, 0)
	}
	return bookings, nil
}

// GetOwnerBookingStats calculates total bookings and revenue for an owner
func GetOwnerBookingStats(ownerID int64, venueID int64) (int64, float64, error) {
	// FIX: Count confirmed, present, AND absent bookings
	query := `
		SELECT 
			COUNT(b.id), 
			COALESCE(SUM(b.total_price), 0)
		FROM bookings b
		JOIN venues v ON b.venue_id = v.id
		WHERE v.owner_id = ? 
        AND b.status IN ('confirmed', 'present', 'absent') 
        AND v.id = ?
	`
	
	var totalBookings int64
	var totalRevenue float64
	
	err := db.DB.QueryRow(query, ownerID, venueID).Scan(&totalBookings, &totalRevenue)
	if err != nil {
		log.Println("Error calculating owner stats:", err)
		return 0, 0, err
	}
	
	return totalBookings, totalRevenue, nil
}

// GetOwnerPopularTime finds the most frequently booked start hour for an owner
func GetOwnerPopularTime(ownerID int64, venueID int64) (string, error) {
	query := `
		SELECT 
			HOUR(CONVERT_TZ(b.start_time, '+00:00', 'Asia/Kolkata')) as popular_hour, 
			COUNT(b.id) as booking_count
		FROM bookings b
		JOIN venues v ON b.venue_id = v.id
		WHERE v.owner_id = ? AND b.status = 'confirmed' AND v.id = ?
		GROUP BY popular_hour
		ORDER BY booking_count DESC
		LIMIT 1
	`
	
	var popularHour sql.NullInt64
	var count int
	
	err := db.DB.QueryRow(query, ownerID, venueID).Scan(&popularHour, &count)
	if err != nil {
		if err == sql.ErrNoRows {
			return "--:--", nil 
		}
		log.Println("Error calculating popular time:", err)
		return "", err
	}

	if !popularHour.Valid {
		return "--:--", nil
	}

	popularTime := time.Date(0, 1, 1, int(popularHour.Int64), 0, 0, 0, time.UTC)
	return popularTime.Format("03:04 PM"), nil
}

// GetPlatformBookingStats calculates total bookings and revenue for the whole platform
func GetPlatformBookingStats() (int64, float64, error) {
	query := `
		SELECT 
			COUNT(id), 
			COALESCE(SUM(total_price), 0)
		FROM bookings
		WHERE status = 'confirmed'
	`
	
	var totalBookings int64
	var totalRevenue float64
	
	err := db.DB.QueryRow(query).Scan(&totalBookings, &totalRevenue)
	if err != nil {
		log.Println("Error calculating platform stats:", err)
		return 0, 0, err
	}
	
	return totalBookings, totalRevenue, nil
}

// GetPlatformPopularTime finds the most frequently booked start hour on the platform
func GetPlatformPopularTime() (string, error) {
	query := `
		SELECT 
			HOUR(CONVERT_TZ(start_time, '+00:00', 'Asia/Kolkata')) as popular_hour, 
			COUNT(id) as booking_count
		FROM bookings
		WHERE status = 'confirmed'
		GROUP BY popular_hour
		ORDER BY booking_count DESC
		LIMIT 1
	`
	
	var popularHour sql.NullInt64
	var count int
	
	err := db.DB.QueryRow(query).Scan(&popularHour, &count)
	if err != nil {
		if err == sql.ErrNoRows {
			return "--:--", nil
		}
		log.Println("Error calculating platform popular time:", err)
		return "", err
	}

	if !popularHour.Valid {
		return "--:--", nil
	}

	popularTime := time.Date(0, 1, 1, int(popularHour.Int64), 0, 0, 0, time.UTC)
	return popularTime.Format("03:04 PM"), nil
}

// GetVenueStatsGrouped calculates total bookings and revenue for all venues (for Admin)
func GetVenueStatsGrouped() ([]VenueStats, error) {
	query := `
		SELECT 
			v.id,
			v.name,
			v.sport_category,
			COUNT(b.id) as total_bookings,
			COALESCE(SUM(b.total_price), 0) as total_revenue
		FROM venues v
		LEFT JOIN bookings b ON v.id = b.venue_id AND b.status = 'confirmed'
		WHERE v.status = 'approved'
		GROUP BY v.id, v.name, v.sport_category
		ORDER BY v.sport_category, total_revenue DESC
	`
	
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Println("Error calculating grouped venue stats:", err)
		return nil, err
	}
	defer rows.Close()

	var statsList []VenueStats
	for rows.Next() {
		var stats VenueStats
		if err := rows.Scan(
			&stats.VenueID,
			&stats.VenueName,
			&stats.SportCategory,
			&stats.TotalBookings,
			&stats.TotalRevenue,
		); err != nil {
			log.Println("Error scanning venue stats:", err)
			continue
		}
		statsList = append(statsList, stats)
	}
	
	if statsList == nil {
		statsList = make([]VenueStats, 0)
	}
	return statsList, nil
}

// GetOwnerVenueStatsGrouped calculates stats per venue for a specific owner
func GetOwnerVenueStatsGrouped(ownerID int64) ([]VenueStats, error) {
	query := `
		SELECT 
			v.id,
			v.name,
			v.sport_category,
			COUNT(b.id) as total_bookings,
			COALESCE(SUM(b.total_price), 0) as total_revenue
		FROM venues v
		LEFT JOIN bookings b ON v.id = b.venue_id AND b.status = 'confirmed'
		WHERE v.owner_id = ?
		GROUP BY v.id, v.name, v.sport_category
	`
	
	rows, err := db.DB.Query(query, ownerID)
	if err != nil {
		log.Println("Error calculating owner grouped stats:", err)
		return nil, err
	}
	defer rows.Close()

	var statsList []VenueStats
	for rows.Next() {
		var stats VenueStats
		if err := rows.Scan(
			&stats.VenueID,
			&stats.VenueName,
			&stats.SportCategory,
			&stats.TotalBookings,
			&stats.TotalRevenue,
		); err != nil {
			log.Println("Error scanning venue stats:", err)
			continue
		}
		statsList = append(statsList, stats)
	}
	
	if statsList == nil {
		statsList = make([]VenueStats, 0)
	}
	return statsList, nil
}

// ConfirmBookingPayment updates status to 'confirmed' after payment
func ConfirmBookingPayment(bookingID int64) error {
	query := `UPDATE bookings SET status = 'confirmed' WHERE id = ?`
	_, err := db.DB.Exec(query, bookingID)
	if err != nil {
		log.Println("Error confirming payment:", err)
		return err
	}
	return nil
}

// FindBookingByID fetches a single booking by its ID
func FindBookingByID(bookingID int64) (*Booking, error) {
	query := `
		SELECT id, user_id, venue_id, start_time, end_time, total_price, status, created_at
		FROM bookings
		WHERE id = ?
	`
	var b Booking
	err := db.DB.QueryRow(query, bookingID).Scan(
		&b.ID, &b.UserID, &b.VenueID, &b.StartTime, &b.EndTime, 
		&b.TotalPrice, &b.Status, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBookedSlotsForDate fetches confirmed OR present bookings
func GetBookedSlotsForDate(venueID int64, dateStr string) ([]BookedSlot, error) {
	// FIX: Updated query to include 'present' status
	query := `
		SELECT start_time, end_time 
		FROM bookings 
		WHERE venue_id = ? 
		AND DATE(start_time) = ? 
		AND status IN ('confirmed', 'present')
	`
	
	rows, err := db.DB.Query(query, venueID, dateStr)
	if err != nil {
		log.Println("Error querying booked slots:", err)
		return nil, err
	}
	defer rows.Close()

	var slots []BookedSlot
	for rows.Next() {
		var s BookedSlot
		if err := rows.Scan(&s.StartTime, &s.EndTime); err != nil {
			continue
		}
		slots = append(slots, s)
	}
	
	if slots == nil {
		slots = make([]BookedSlot, 0)
	}
	return slots, nil
}

// UpdateBookingStatusByOwner updates status if the requester owns the venue
func UpdateBookingStatusByOwner(bookingID int64, ownerID int64, newStatus string) error {
	// We use a JOIN to verify the link between Booking -> Venue -> Owner
	query := `
		UPDATE bookings b
		JOIN venues v ON b.venue_id = v.id
		SET b.status = ?
		WHERE b.id = ? AND v.owner_id = ?
	`
	
	result, err := db.DB.Exec(query, newStatus, bookingID, ownerID)
	if err != nil {
		log.Println("Error updating booking by owner:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("booking not found or you do not own this venue")
	}
	return nil
}
// UpdateBookingStatusDirect allows admins to update status without ownership check
func UpdateBookingStatusDirect(bookingID int64, newStatus string) error {
	query := `UPDATE bookings SET status = ? WHERE id = ?`
	_, err := db.DB.Exec(query, newStatus, bookingID)
	if err != nil {
		log.Println("Error updating booking directly:", err)
		return err
	}
	return nil
}

// booking/booking_repository.go

// GetVenueStatsSimple calculates stats for a venue (No owner check - for Admin)
func GetVenueStatsSimple(venueID int64) (int64, float64, error) {
	query := `
		SELECT 
			COUNT(id), 
			COALESCE(SUM(total_price), 0)
		FROM bookings
		WHERE venue_id = ? 
        AND status IN ('confirmed', 'present', 'absent')
	`
	
	var totalBookings int64
	var totalRevenue float64
	
	err := db.DB.QueryRow(query, venueID).Scan(&totalBookings, &totalRevenue)
	if err != nil {
		log.Println("Error calculating venue stats:", err)
		return 0, 0, err
	}
	
	return totalBookings, totalRevenue, nil
}

// GetVenuePopularTimeSimple finds popular time for a venue (No owner check - for Admin)
func GetVenuePopularTimeSimple(venueID int64) (string, error) {
	query := `
		SELECT 
			HOUR(CONVERT_TZ(start_time, '+00:00', 'Asia/Kolkata')) as popular_hour, 
			COUNT(id) as booking_count
		FROM bookings
		WHERE venue_id = ? 
        AND status IN ('confirmed', 'present', 'absent')
		GROUP BY popular_hour
		ORDER BY booking_count DESC
		LIMIT 1
	`
	
	var popularHour sql.NullInt64 
	var count int
	
	err := db.DB.QueryRow(query, venueID).Scan(&popularHour, &count)
	if err != nil {
		if err == sql.ErrNoRows {
			return "--:--", nil
		}
		log.Println("Error calculating popular time:", err)
		return "", err
	}

	if !popularHour.Valid {
		return "--:--", nil
	}

	popularTime := time.Date(0, 1, 1, int(popularHour.Int64), 0, 0, 0, time.UTC)
	return popularTime.Format("03:04 PM"), nil
}