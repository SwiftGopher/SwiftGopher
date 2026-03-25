package modules

type TransportType string

const (
	TransportBike    TransportType = "bike"
	TransportCar     TransportType = "car"
	TransportFoot    TransportType = "foot"
	TransportScooter TransportType = "scooter"
)

type CourierStatus string

const (
	StatusFree    CourierStatus = "free"
	StatusBusy    CourierStatus = "busy"
	StatusOffline CourierStatus = "offline"
)

type Courier struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	TransportType TransportType `json:"transport_type"`
	Status        CourierStatus `json:"status"`
	CurrentLat    float64       `json:"current_lat"`
	CurrentLng    float64       `json:"current_lng"`
}

type UpdateStatusRequest struct {
	Status CourierStatus `json:"status"`
}
