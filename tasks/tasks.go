package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Task type constants
const (
	TypeEmailDelivery = "email:delivery"
	TypeDataProcess   = "data:process"
)

// EmailPayload represents the payload for email tasks
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// DataProcessPayload represents the payload for data processing tasks
type DataProcessPayload struct {
	DataID string `json:"data_id"`
	Action string `json:"action"`
	Delay  int    `json:"delay"` // Delay in seconds to simulate processing
}

// NewEmailDeliveryTask creates a new email delivery task
func NewEmailDeliveryTask(to, subject, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailDelivery, payload), nil
}

// NewDataProcessTask creates a new data processing task
func NewDataProcessTask(dataID, action string, delay int) (*asynq.Task, error) {
	payload, err := json.Marshal(DataProcessPayload{
		DataID: dataID,
		Action: action,
		Delay:  delay,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDataProcess, payload), nil
}

// HandleEmailDeliveryTask handles email delivery tasks
func HandleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("📧 [Email Task] Sending email to: %s", p.To)
	log.Printf("   Subject: %s", p.Subject)
	log.Printf("   Body: %s", p.Body)

	// Simulate email sending
	time.Sleep(1 * time.Second)

	log.Printf("✅ [Email Task] Successfully sent email to: %s", p.To)
	return nil
}

// HandleDataProcessTask handles data processing tasks
func HandleDataProcessTask(ctx context.Context, t *asynq.Task) error {
	var p DataProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("🔄 [Data Process Task] Processing data: %s", p.DataID)
	log.Printf("   Action: %s", p.Action)
	log.Printf("   Simulating processing for %d seconds...", p.Delay)

	// Simulate data processing with configurable delay
	time.Sleep(time.Duration(p.Delay) * time.Second)

	log.Printf("✅ [Data Process Task] Successfully processed data: %s", p.DataID)
	return nil
}
