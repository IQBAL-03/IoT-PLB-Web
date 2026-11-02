package model

// HistoryEntry represents a single pump activation or sensor event record.
type HistoryEntry struct {
	ID        int     `json:"id"`
	DeviceID  string  `json:"device_id"`
	Timestamp string  `json:"timestamp"`
	Action    string  `json:"action"`
	Mode      string  `json:"mode"`
	Moisture  float64 `json:"moisture"`
	Duration  string  `json:"duration"`
}

// WebhookPayload represents telemetry data sent by ESP32 or Blynk Webhook.
type WebhookPayload struct {
	DeviceID string  `json:"device_id"`
	Moisture float64 `json:"moisture"`
	PumpOn   *bool   `json:"pump_on,omitempty"`
	Action   string  `json:"action,omitempty"`
	Mode     string  `json:"mode,omitempty"`
	ForceLog bool    `json:"force_log,omitempty"`
}

