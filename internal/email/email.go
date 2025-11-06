package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/autofileingest/internal/config"
	"github.com/autofileingest/internal/transfer"
)

// Notifier handles email notifications
type Notifier struct {
	config *config.Config
}

// NewNotifier creates a new email notifier
func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{
		config: cfg,
	}
}

// SendTransferComplete sends a notification when transfer is complete
func (n *Notifier) SendTransferComplete(deviceName string, stats transfer.TransferStats, logPath string) error {
	if !n.config.Email.Enabled {
		return nil
	}

	subject := strings.ReplaceAll(n.config.Email.Subject, "{device}", deviceName)
	body := n.buildEmailBody(deviceName, stats)

	return n.sendEmail(subject, body, logPath)
}

// buildEmailBody creates the email body content
func (n *Notifier) buildEmailBody(deviceName string, stats transfer.TransferStats) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Media Ingest Complete - %s\n", deviceName))
	buf.WriteString(strings.Repeat("=", 50) + "\n\n")

	buf.WriteString(fmt.Sprintf("Transfer Summary:\n"))
	buf.WriteString(fmt.Sprintf("  Device: %s\n", deviceName))
	buf.WriteString(fmt.Sprintf("  Total Files: %d\n", stats.TotalFiles))
	buf.WriteString(fmt.Sprintf("  Successfully Transferred: %d\n", stats.ProcessedFiles-stats.FailedFiles))
	buf.WriteString(fmt.Sprintf("  Failed: %d\n", stats.FailedFiles))
	buf.WriteString(fmt.Sprintf("  Skipped: %d\n", stats.SkippedFiles))
	buf.WriteString(fmt.Sprintf("  Total Size: %s\n", formatBytes(stats.TotalBytes)))
	buf.WriteString(fmt.Sprintf("  Duration: %s\n", time.Since(stats.StartTime).Round(time.Second)))
	buf.WriteString(fmt.Sprintf("  Average Speed: %s/s\n", formatBytes(int64(stats.TransferredBytes/int64(time.Since(stats.StartTime).Seconds())))))

	buf.WriteString("\n")
	buf.WriteString("This is an automated message from Media Ingest Server.\n")

	return buf.String()
}

// sendEmail sends an email using SMTP
func (n *Notifier) sendEmail(subject, body, attachment string) error {
	// Build email message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", n.config.Email.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(n.config.Email.To, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// TODO: Add attachment support if needed

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", n.config.Email.SMTPHost, n.config.Email.SMTPPort)
	auth := smtp.PlainAuth("", n.config.Email.Username, n.config.Email.Password, n.config.Email.SMTPHost)

	// Send email
	err := smtp.SendMail(addr, auth, n.config.Email.From, n.config.Email.To, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// formatBytes formats bytes as human-readable size
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}
