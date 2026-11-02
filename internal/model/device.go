package model

// DeviceItem represents an IoT unit registered in the system.
type DeviceItem struct {
	DeviceID    string `json:"device_id"`
	Name        string `json:"name"`
	IsOwner     bool   `json:"is_owner"`
	HasOwner    bool   `json:"has_owner"`
	PairingCode string `json:"pairing_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ClaimReq defines payload to claim ownership of an IoT device.
type ClaimReq struct {
	DeviceID    string `json:"device_id"`
	PairingCode string `json:"pairing_code"`
	Name        string `json:"name"`
}

