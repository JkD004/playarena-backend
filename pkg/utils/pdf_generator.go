package utils

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

func GenerateTicketPDF(
	userName string,
	userLastName string,
	venueName string,
	venueAddress string,
	bookingID int64,
	userID int64,
	startTime time.Time,
	endTime time.Time,
	bookingCreated time.Time,
) (*gofpdf.Fpdf, error) {

	loc, _ := time.LoadLocation("Asia/Kolkata")
	startTime = startTime.In(loc)
	endTime = endTime.In(loc)
	bookingCreated = bookingCreated.In(loc)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	// COLORS
	dark := []int{17, 24, 39}
	gray := []int{107, 114, 128}
	lightGray := []int{243, 244, 246}
	border := []int{229, 231, 235}
	green := []int{21, 128, 61}
	greenBg := []int{220, 252, 231}

	cardX, cardY, cardW := 25.0, 20.0, 160.0

	// CARD
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(cardX, cardY, cardW, 245, "F")

	// HEADER
	pdf.SetFillColor(dark[0], dark[1], dark[2])
	pdf.Rect(cardX, cardY, cardW, 32, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetXY(cardX, cardY+10)
	pdf.CellFormat(cardW, 8, "SPORTGRID", "", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(cardX, cardY+20)
	pdf.CellFormat(cardW, 5, "OFFICIAL ENTRY PASS", "", 0, "C", false, 0, "")

	// BOOKED ON
	y := cardY + 38
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX+10, y)
	pdf.Cell(40, 5, "BOOKED ON")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(cardX+10, y+6)
	pdf.Cell(40, 6, bookingCreated.Format("02 Jan 2006, 03:04 PM"))

	// STATUS
	pdf.SetFillColor(greenBg[0], greenBg[1], greenBg[2])
	pdf.SetTextColor(green[0], green[1], green[2])
	pdf.SetFont("Arial", "B", 9)
	pdf.RoundedRect(cardX+cardW-45, y, 35, 10, 2, "F", "1234")
	pdf.SetXY(cardX+cardW-45, y+2)
	pdf.CellFormat(35, 6, "CONFIRMED", "", 0, "C", false, 0, "")

	// DIVIDER
	pdf.SetDrawColor(border[0], border[1], border[2])
	pdf.Line(cardX+10, y+16, cardX+cardW-10, y+16)

	// VENUE
	y += 22
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.SetXY(cardX+10, y)
	pdf.Cell(0, 8, venueName)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX+10, y+8)
	pdf.Cell(0, 6, venueAddress)

	// PLAYER + IDS
	y += 24
	pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.RoundedRect(cardX+10, y, cardW-20, 36, 4, "F", "1234")

	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX+16, y+6)
	pdf.Cell(40, 4, "PLAYER")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.SetXY(cardX+16, y+12)
	pdf.Cell(40, 6, userName+" "+userLastName)

	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX+90, y+6)
	pdf.Cell(40, 4, "PLAYER ID")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.SetXY(cardX+90, y+12)
	pdf.Cell(40, 6, fmt.Sprintf("#%d", userID))

	// Ticket + Booking
	pdf.SetXY(cardX+16, y+22)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.Cell(40, 4, "TICKET ID")

	pdf.SetXY(cardX+16, y+28)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.Cell(40, 6, fmt.Sprintf("#%d", bookingID))

	pdf.SetXY(cardX+90, y+22)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.Cell(40, 4, "BOOKING ID")

	pdf.SetXY(cardX+90, y+28)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.Cell(40, 6, fmt.Sprintf("#%d", bookingID))

	// SLOT
	y += 50
	pdf.SetDrawColor(border[0], border[1], border[2])
	pdf.RoundedRect(cardX+10, y, cardW-20, 36, 6, "D", "1234")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX+10, y+6)
	pdf.CellFormat(cardW-20, 5, "SCHEDULED SLOT", "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(15, 118, 110)
	pdf.SetXY(cardX+10, y+14)
	pdf.CellFormat(cardW-20, 7, startTime.Format("02 Jan 2006"), "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(dark[0], dark[1], dark[2])
	pdf.SetXY(cardX+10, y+22)
	pdf.CellFormat(cardW-20, 8,
		startTime.Format("03:04 PM")+" - "+endTime.Format("03:04 PM"),
		"", 0, "C", false, 0, "",
	)

	// QR
	qrY := y + 44
	qr, _ := qrcode.Encode(fmt.Sprintf("BOOKING:%d", bookingID), qrcode.Medium, 256)
	pdf.RegisterImageOptionsReader("qr", gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(qr))
	pdf.ImageOptions("qr", cardX+(cardW/2)-15, qrY, 30, 30, false, gofpdf.ImageOptions{}, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(gray[0], gray[1], gray[2])
	pdf.SetXY(cardX, qrY+34)
	pdf.CellFormat(cardW, 5, "Scan to verify entry", "", 0, "C", false, 0, "")

	return pdf, nil
}
