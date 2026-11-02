package model

// SystemStatus represents the real-time operational status of an IoT unit.
type SystemStatus struct {
	DeviceID  string  `json:"device_id"`
	PumpOn    bool    `json:"pump_on"`
	Moisture  float64 `json:"moisture"`
	Mode      string  `json:"mode"`
	Threshold float64 `json:"threshold"`
	Timestamp string  `json:"timestamp"`
}

// PumpReq defines the request body for turning pump ON/OFF.
type PumpReq struct {
	DeviceID string `json:"device_id"`
}

// ModeReq defines payload to switch operation mode between 'auto' and 'manual'.
type ModeReq struct {
	DeviceID string `json:"device_id"`
	Mode     string `json:"mode"`
}

