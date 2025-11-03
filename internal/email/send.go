package email

import (
	"log"
	"time"
)

func SendEmail(to, subject, body string) error {

	log.Printf("[SIMULATION] Sending email to %s\n", to)
	time.Sleep(200 * time.Millisecond)
	log.Printf("[SIMULATION] Email sent successfully to %s\n", to)
	return nil
}
