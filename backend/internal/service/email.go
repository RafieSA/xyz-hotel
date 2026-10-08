package service

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var emailMu sync.Mutex

// logPath returns the email log file path relative to backend working dir.
// In production binary runs from backend/ so logs/email.log is correct.
// Fallback to backend/logs/email.log if logs dir not present at CWD.
func logPath() string {
	// prefer backend/logs/email.log when running from repo root
	candidates := []string{
		filepath.Join("logs", "email.log"),
		filepath.Join("backend", "logs", "email.log"),
	}
	for _, p := range candidates {
		dir := filepath.Dir(p)
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			return p
		}
	}
	return candidates[0]
}

// Send writes a log-only email entry to backend/logs/email.log.
// It creates the logs directory if needed, appends a timestamped entry,
// and logs via slog. Never uses SMTP.
func Send(to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		to = "unknown@xyz-hotel.local"
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "(no subject)"
	}
	path := logPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("email log mkdir failed", "err", err, "dir", dir)
		return err
	}

	now := time.Now().Format(time.RFC3339)
	entry := strings.Builder{}
	entry.WriteString("========================================\n")
	entry.WriteString(fmt.Sprintf("Time: %s\n", now))
	entry.WriteString(fmt.Sprintf("To: %s\n", to))
	entry.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	entry.WriteString("Body:\n")
	entry.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		entry.WriteString("\n")
	}
	entry.WriteString("========================================\n\n")

	emailMu.Lock()
	defer emailMu.Unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("email log open failed", "err", err, "path", path)
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(entry.String()); err != nil {
		slog.Error("email log write failed", "err", err)
		return err
	}
	slog.Info("email logged", "to", to, "subject", subject, "path", path)
	return nil
}

// SendAsync fires Send in a non-blocking goroutine. Errors are logged via slog.
func SendAsync(to, subject, body string) {
	go func() {
		if err := Send(to, subject, body); err != nil {
			slog.Error("async email log failed", "err", err, "to", to, "subject", subject)
		}
	}()
}

// Template helpers: warm, clear, hospitality tone with hierarchy.

func BookingCreatedSubject(bookingID int64) string {
	return fmt.Sprintf("Your XYZ Hotel booking #%d is received - pending payment", bookingID)
}

func BookingCreatedBody(bookingID, roomTypeID int64, checkIn, checkOut string, guests int, totalPrice int64) string {
	return fmt.Sprintf(`Dear Guest,

Warm greetings from XYZ Hotel!

Your booking has been received and is now awaiting payment.

  Booking ID   : #%d
  Room Type ID : %d
  Check-in     : %s
  Check-out    : %s
  Guests       : %d
  Total Price  : Rp %d

Next steps:
  1. Please upload your payment proof within 12 hours via My Bookings.
  2. Our team will verify your payment shortly after upload.
  3. You will receive a confirmation email once verified.

If you have any questions, please contact our reception at +62 812-3456-7890.

We look forward to welcoming you!

Warm regards,
XYZ Hotel Team
Jl. Kenangan No. 8, Jakarta`, bookingID, roomTypeID, checkIn, checkOut, guests, totalPrice)
}

func BookingVerifiedSubject(bookingID int64) string {
	return fmt.Sprintf("Your XYZ Hotel booking #%d is verified - see you soon!", bookingID)
}

func BookingVerifiedBody(bookingID int64, checkIn, checkOut string) string {
	return fmt.Sprintf(`Dear Guest,

Great news! Your booking #%d has been verified.

  Check-in  : %s
  Check-out : %s

Your reservation is now confirmed. Please bring your ID at check-in.
Check-in time starts at 14:00, check-out by 12:00.

We cannot wait to host your stay. If you need airport pickup or extra amenities, reply to this email.

Warm regards,
XYZ Hotel Team`, bookingID, checkIn, checkOut)
}

func BookingRejectedSubject(bookingID int64) string {
	return fmt.Sprintf("Your XYZ Hotel booking #%d needs attention", bookingID)
}

func BookingRejectedBody(bookingID int64, reason string) string {
	if strings.TrimSpace(reason) == "" {
		reason = "Payment proof could not be verified. Please re-upload a clearer image or contact reception."
	}
	return fmt.Sprintf(`Dear Guest,

We reviewed your booking #%d and could not verify your payment.

  Reason: %s

What to do next:
  - Re-upload a clearer payment proof, or
  - Contact reception at +62 812-3456-7890 for manual assistance.

We are here to help you complete your reservation.

Warm regards,
XYZ Hotel Team`, bookingID, reason)
}

func BookingExpiredSubject(bookingID int64) string {
	return fmt.Sprintf("Your XYZ Hotel booking #%d has expired", bookingID)
}

func BookingExpiredBody(bookingID int64) string {
	return fmt.Sprintf(`Dear Guest,

Your booking #%d has expired because we did not receive payment proof within 12 hours.

No charges were made. Your selected dates have been released for other guests.
You are welcome to create a new booking at any time - we would love to host you.

If you believe this is a mistake or you uploaded proof recently, please contact us at +62 812-3456-7890.

Warm regards,
XYZ Hotel Team`, bookingID)
}

func BookingCancelledSubject(bookingID int64) string {
	return fmt.Sprintf("Your XYZ Hotel booking #%d has been cancelled", bookingID)
}

func BookingCancelledBody(bookingID int64) string {
	return fmt.Sprintf(`Dear Guest,

Your booking #%d has been cancelled as requested.

If a voucher was applied, its usage has been logged for review - please contact reception if you need re-issuance.
We hope to welcome you another time. Your next stay is just a booking away.

Warm regards,
XYZ Hotel Team`, bookingID)
}

func BookingCheckedInSubject(bookingID int64) string {
	return fmt.Sprintf("Welcome! Booking #%d checked in", bookingID)
}

func BookingCheckedInBody(bookingID int64, unitCode string) string {
	return fmt.Sprintf(`Dear Guest,

Welcome to XYZ Hotel! You have been checked in.

  Booking : #%d
  Room    : %s

Enjoy your stay - WiFi password is at the front desk. Dial 0 for assistance.

Warm regards,
XYZ Hotel Team`, bookingID, unitCode)
}

func BookingCheckedOutSubject(bookingID int64) string {
	return fmt.Sprintf("Thank you for staying - booking #%d checked out", bookingID)
}

func BookingCheckedOutBody(bookingID int64) string {
	return fmt.Sprintf(`Dear Guest,

Thank you for staying with XYZ Hotel! Booking #%d is now checked out.

We would love your feedback - please leave a review on our website to help future guests.
Safe travels, and we hope to see you again soon.

Warm regards,
XYZ Hotel Team`, bookingID)
}
