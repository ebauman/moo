package types

import "time"

var (
	StatusUnknown  Status = "unknown"
	StatusAccepted Status = "accepted"
	StatusHeld     Status = "held"
	StatusDenied   Status = "denied"
	StatusPending  Status = "pending"
	StatusError    Status = "error"
)

type Agent struct {
	ID            string    `json:"id"`
	Secret        string    `json:"secret"`
	IP            string    `json:"ip"`
	Status        Status    `json:"status"`
	ManifestUrl   string    `json:"manifestUrl"`
	StatusMessage string    `json:"statusMessage"`
	Completed     bool      `json:"completed"`
	LastContact   time.Time `json:"time"`

	ClusterName string
	UseExisting bool
}

type Status string
