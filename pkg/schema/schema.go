package schema

import "time"

type HealthResponse struct {
	R    string    `json:"r"`
	Time time.Time `json:"time"`
}
