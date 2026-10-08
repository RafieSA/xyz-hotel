package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jung-kurt/gofpdf"
)

// InvoiceData holds joined data for PDF rendering.
type InvoiceData struct {
	BookingID    int64
	Status       string
	CheckIn      time.Time
	CheckOut     time.Time
	Guests       int
	TotalPrice   int64
	ProofURL     *string
	VoucherID    *int64
	CreatedAt    time.Time
	UserName     string
	UserEmail    string
	RoomTypeName string
	RoomTypeDesc string
	RoomCapacity int
	PricePerNight int64
	VoucherCode   *string
	VoucherDiscount *float64
}

// InvoiceService generates premium WarmAura PDFs.
type InvoiceService struct {
	DB *sqlx.DB
}

func NewInvoiceService(db *sqlx.DB) *InvoiceService {
	return &InvoiceService{DB: db}
}

// GenerateInvoice queries booking + user + room_type (+ voucher) and creates PDF at storage/invoices/invoice-{id}.pdf
func (s *InvoiceService) GenerateInvoice(ctx context.Context, bookingID int64) (string, error) {
	if s.DB == nil {
		return "", fmt.Errorf("database not available")
	}
	var d InvoiceData
	query := `
		SELECT
			b.id as booking_id, b.status, b.check_in, b.check_out, b.guests, b.total_price, b.proof_url, b.voucher_id, b.created_at,
			u.name as user_name, u.email as user_email,
			rt.name as room_type_name, rt.description as room_type_desc, rt.capacity as room_capacity, rt.price as price_per_night,
			v.code as voucher_code, v.discount_percent as voucher_discount
		FROM bookings b
		JOIN users u ON u.id = b.user_id
		JOIN room_types rt ON rt.id = b.room_type_id
		LEFT JOIN vouchers v ON v.id = b.voucher_id
		WHERE b.id = $1
	`
	// Use sql.Null types to allow LEFT JOIN nulls
	var row struct {
		BookingID       int64          `db:"booking_id"`
		Status          string         `db:"status"`
		CheckIn         time.Time      `db:"check_in"`
		CheckOut        time.Time      `db:"check_out"`
		Guests          int            `db:"guests"`
		TotalPrice      int64          `db:"total_price"`
		ProofURL        sql.NullString `db:"proof_url"`
		VoucherID       sql.NullInt64  `db:"voucher_id"`
		CreatedAt       time.Time      `db:"created_at"`
		UserName        string         `db:"user_name"`
		UserEmail       string         `db:"user_email"`
		RoomTypeName    string         `db:"room_type_name"`
		RoomTypeDesc    string         `db:"room_type_desc"`
		RoomCapacity    int            `db:"room_capacity"`
		PricePerNight   int64          `db:"price_per_night"`
		VoucherCode     sql.NullString `db:"voucher_code"`
		VoucherDiscount sql.NullFloat64 `db:"voucher_discount"`
	}
	if err := s.DB.GetContext(ctx, &row, query, bookingID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("booking not found")
		}
		return "", fmt.Errorf("failed to load booking: %w", err)
	}
	d = InvoiceData{
		BookingID:     row.BookingID,
		Status:        row.Status,
		CheckIn:       row.CheckIn,
		CheckOut:      row.CheckOut,
		Guests:        row.Guests,
		TotalPrice:    row.TotalPrice,
		CreatedAt:     row.CreatedAt,
		UserName:      row.UserName,
		UserEmail:     row.UserEmail,
		RoomTypeName:  row.RoomTypeName,
		RoomTypeDesc:  row.RoomTypeDesc,
		RoomCapacity:  row.RoomCapacity,
		PricePerNight: row.PricePerNight,
	}
	if row.ProofURL.Valid {
		d.ProofURL = &row.ProofURL.String
	}
	if row.VoucherID.Valid {
		v := row.VoucherID.Int64
		d.VoucherID = &v
	}
	if row.VoucherCode.Valid {
		v := row.VoucherCode.String
		d.VoucherCode = &v
	}
	if row.VoucherDiscount.Valid {
		v := row.VoucherDiscount.Float64
		d.VoucherDiscount = &v
	}

	return s.renderPDF(d)
}

func formatIDR(n int64) string {
	// format 1250000 -> Rp 1.250.000
	s := fmt.Sprintf("%d", n)
	// insert dots every 3 from right
	var out string
	for i, c := range s {
		remaining := len(s) - i
		if i > 0 && remaining%3 == 0 {
			needSep := true
			// only insert if not at boundary of first digit grouping? simpler: insert when position
			if len(out) > 0 {
				_ = needSep
			}
		}
		_ = c
		_ = remaining
		_ = out
		break
	}
	// simpler correct implementation
	var formatted string
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if count == 3 {
			formatted = "." + formatted
			count = 0
		}
		formatted = string(s[i]) + formatted
		count++
	}
	return "Rp " + formatted
}

func (s *InvoiceService) renderPDF(d InvoiceData) (string, error) {
	dir := filepath.Join("storage", "invoices")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create invoice dir: %w", err)
	}
	filename := fmt.Sprintf("invoice-%d.pdf", d.BookingID)
	fullPath := filepath.Join(dir, filename)

	nights := int(d.CheckOut.Sub(d.CheckIn).Hours() / 24)
	if nights < 1 {
		nights = 1
	}
	subtotal := d.PricePerNight * int64(nights)
	discountAmt := subtotal - d.TotalPrice
	if discountAmt < 0 {
		discountAmt = 0
	}
	discountPct := 0.0
	if d.VoucherDiscount != nil {
		discountPct = *d.VoucherDiscount
	} else if subtotal > 0 && discountAmt > 0 {
		discountPct = float64(discountAmt) / float64(subtotal) * 100
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 0, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Colors
	// WarmAura header #8B5A2B -> 139,90,43
	// Cream text #FFF8E7 -> 255,248,231
	// Gold line #D4A574 -> 212,165,116
	// Dark brown text #3E2723 -> 62,39,35
	// Muted #8D6E63 -> 141,110,99
	// Light cream bg #FFFBF5 -> 255,251,245
	// Border #E8DCC8 -> 232,220,200

	pageW, _ := pdf.GetPageSize()
	_ = pageW

	// Header background
	pdf.SetFillColor(139, 90, 43)
	pdf.Rect(0, 0, 210, 38, "F")
	// Gold accent line at bottom of header
	pdf.SetFillColor(212, 165, 116)
	pdf.Rect(0, 38, 210, 1.2, "F")

	// Header content
	pdf.SetY(9)
	pdf.SetX(15)
	pdf.SetTextColor(255, 248, 231)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 8, "XYZ HOTEL", "", 0, "L", false, 0, "")
	// Invoice title right aligned
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetXY(130, 11)
	pdf.SetTextColor(255, 248, 231)
	pdf.CellFormat(65, 5, "INVOICE", "", 0, "R", false, 0, "")
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(232, 220, 200)
	pdf.SetXY(130, 17)
	pdf.CellFormat(65, 4, "Premium Hospitality  |  xyz-hotel.com", "", 0, "R", false, 0, "")

	pdf.SetXY(15, 24)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(255, 248, 231)
	pdf.CellFormat(0, 4, "Jl. WarmAura No. 88, Bandung 40111  |  hello@xyz-hotel.com  |  +62 22 1234 5678", "", 0, "L", false, 0, "")

	// Move below header
	pdf.SetY(46)
	pdf.SetTextColor(62, 39, 35)

	// Invoice meta row: Invoice # + Dates badge
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(62, 39, 35)
	pdf.CellFormat(0, 7, fmt.Sprintf("Invoice #%06d", d.BookingID), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(141, 110, 99)
	pdf.CellFormat(0, 4, fmt.Sprintf("Issued %s  |  Status %s", d.CreatedAt.Format("02 Jan 2006 15:04 WIB"), d.Status), "", 1, "L", false, 0, "")

	// Divider
	pdf.Ln(3)
	pdf.SetDrawColor(232, 220, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)

	// Two columns: Guest Info | Booking Details
	yStart := pdf.GetY()
	leftX := 15.0
	rightX := 110.0
	colW := 85.0

	// Guest card background
	pdf.SetFillColor(255, 251, 245)
	pdf.SetDrawColor(232, 220, 200)
	pdf.Rect(leftX, yStart, colW, 38, "DF")
	pdf.Rect(rightX, yStart, colW, 38, "DF")

	// Guest info
	pdf.SetXY(leftX+4, yStart+3)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(139, 90, 43)
	pdf.CellFormat(colW-8, 4, "GUEST INFORMATION", "", 0, "L", false, 0, "")
	pdf.SetXY(leftX+4, yStart+8)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(62, 39, 35)
	pdf.CellFormat(colW-8, 5, d.UserName, "", 0, "L", false, 0, "")
	pdf.SetXY(leftX+4, yStart+14)
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(121, 85, 72)
	pdf.CellFormat(colW-8, 4, d.UserEmail, "", 0, "L", false, 0, "")
	pdf.SetXY(leftX+4, yStart+19)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(141, 110, 99)
	pdf.CellFormat(colW-8, 4, fmt.Sprintf("Guests: %d  |  Capacity: %d", d.Guests, d.RoomCapacity), "", 0, "L", false, 0, "")
	// status badge inside guest card
	pdf.SetXY(leftX+4, yStart+26)
	badgeColorR, badgeColorG, badgeColorB := statusBadgeColor(d.Status)
	pdf.SetFillColor(badgeColorR, badgeColorG, badgeColorB)
	badgeW := pdf.GetStringWidth(d.Status) + 8
	if badgeW < 22 {
		badgeW = 22
	}
	pdf.RoundedRect(leftX+4, yStart+25.5, badgeW, 6, 1.5, "1234", "F")
	pdf.SetXY(leftX+4, yStart+25.5)
	pdf.SetFont("Helvetica", "B", 6)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(badgeW, 6, d.Status, "", 0, "C", false, 0, "")

	// Booking details column
	pdf.SetXY(rightX+4, yStart+3)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(139, 90, 43)
	pdf.CellFormat(colW-8, 4, "BOOKING DETAILS", "", 0, "L", false, 0, "")
	details := [][]string{
		{"Check-in", d.CheckIn.Format("02 Jan 2006")},
		{"Check-out", d.CheckOut.Format("02 Jan 2006")},
		{"Nights", fmt.Sprintf("%d night(s)", nights)},
		{"Room Type", d.RoomTypeName},
	}
	dy := 8.5
	for _, row := range details {
		pdf.SetXY(rightX+4, yStart+dy)
		pdf.SetFont("Helvetica", "", 6.5)
		pdf.SetTextColor(141, 110, 99)
		pdf.CellFormat(22, 4, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 7)
		pdf.SetTextColor(62, 39, 35)
		pdf.CellFormat(colW-30, 4, row[1], "", 0, "L", false, 0, "")
		dy += 5
	}

	pdf.SetY(yStart + 42)

	// Room description subtle
	if d.RoomTypeDesc != "" {
		pdf.SetFont("Helvetica", "I", 7)
		pdf.SetTextColor(141, 110, 99)
		pdf.MultiCell(180, 3.5, d.RoomTypeDesc, "", "L", false)
		pdf.Ln(2)
	}

	// Price breakdown table
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(139, 90, 43)
	pdf.CellFormat(0, 6, "PRICE BREAKDOWN", "", 1, "L", false, 0, "")
	// Table header
	pdf.SetFillColor(139, 90, 43)
	pdf.SetTextColor(255, 248, 231)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(95, 7, "  Description", "0", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, "Qty", "0", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Unit Price", "0", 0, "R", true, 0, "")
	pdf.CellFormat(30, 7, "Amount   ", "0", 1, "R", true, 0, "")

	// Row style helper
	drawRow := func(desc, qty, unitPrice, amount string, fill bool) {
		if fill {
			pdf.SetFillColor(255, 251, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(62, 39, 35)
		pdf.SetFont("Helvetica", "", 7)
		pdf.CellFormat(95, 6, "  "+desc, "0", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, qty, "0", 0, "C", true, 0, "")
		pdf.CellFormat(30, 6, unitPrice, "0", 0, "R", true, 0, "")
		pdf.SetFont("Helvetica", "", 7)
		pdf.CellFormat(30, 6, amount+"   ", "0", 1, "R", true, 0, "")
	}

	drawRow(d.RoomTypeName+"  ("+fmt.Sprintf("%d", d.RoomCapacity)+" guests max)", fmt.Sprintf("%d", nights), formatIDR(d.PricePerNight), formatIDR(subtotal), false)

	if discountAmt > 0 {
		label := "Discount"
		if d.VoucherCode != nil {
			label = fmt.Sprintf("Discount (%s  -%.0f%%)", *d.VoucherCode, discountPct)
		} else if discountPct > 0 {
			label = fmt.Sprintf("Discount (-%.0f%%)", discountPct)
		}
		pdf.SetTextColor(46, 125, 50)
		// draw with green tint
		pdf.SetFillColor(232, 245, 233)
		pdf.SetFont("Helvetica", "", 7)
		pdf.CellFormat(95, 6, "  "+label, "0", 0, "L", true, 0, "")
		pdf.CellFormat(25, 6, "-", "0", 0, "C", true, 0, "")
		pdf.CellFormat(30, 6, "", "0", 0, "R", true, 0, "")
		pdf.CellFormat(30, 6, "-"+formatIDR(discountAmt)+"   ", "0", 1, "R", true, 0, "")
	}

	// Total row
	pdf.SetFillColor(62, 39, 35)
	pdf.SetTextColor(255, 248, 231)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(150, 8, "  TOTAL ", "0", 0, "R", true, 0, "")
	pdf.CellFormat(30, 8, formatIDR(d.TotalPrice)+"   ", "0", 1, "R", true, 0, "")

	// Proof info if exists
	pdf.Ln(3)
	pdf.SetTextColor(141, 110, 99)
	pdf.SetFont("Helvetica", "", 6.5)
	if d.ProofURL != nil && *d.ProofURL != "" {
		pdf.CellFormat(0, 4, "Payment proof: "+*d.ProofURL+"  |  Verified via XYZ Hotel secure storage", "", 1, "L", false, 0, "")
	} else {
		pdf.CellFormat(0, 4, "Payment proof: not yet uploaded", "", 1, "L", false, 0, "")
	}

	// Footer
	footerY := 270.0
	// gold line
	pdf.SetFillColor(212, 165, 116)
	pdf.Rect(15, footerY, 180, 0.6, "F")
	pdf.SetXY(15, footerY+3)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(139, 90, 43)
	pdf.CellFormat(0, 4, "Thank you for choosing XYZ Hotel", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 6.5)
	pdf.SetTextColor(141, 110, 99)
	pdf.CellFormat(0, 3.5, "Jl. WarmAura No. 88, Bandung 40111  |  hello@xyz-hotel.com  |  +62 22 1234 5678  |  www.xyz-hotel.com", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "I", 6)
	pdf.SetTextColor(160, 130, 115)
	pdf.CellFormat(0, 3.5, "This is a computer-generated invoice. No signature required.  |  WarmAura Hospitality  |  Invoice #"+fmt.Sprintf("%06d", d.BookingID), "", 1, "C", false, 0, "")

	if err := pdf.OutputFileAndClose(fullPath); err != nil {
		return "", fmt.Errorf("failed to write pdf: %w", err)
	}
	rel := filepath.Join("storage", "invoices", filename)
	return rel, nil
}

func statusBadgeColor(status string) (int, int, int) {
	switch status {
	case "verified":
		return 46, 125, 50
	case "checked_in":
		return 21, 101, 192
	case "checked_out":
		return 55, 71, 79
	case "pending_payment":
		return 245, 124, 0
	case "waiting_verification":
		return 123, 31, 162
	default:
		return 139, 90, 43
	}
}
