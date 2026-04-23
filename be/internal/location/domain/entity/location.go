package entity

type Location struct {
	ID        string
	Name      string
	Address   string
	Latitude  float64
	Longitude float64
}

type NearbyLocation struct {
	Location       Location
	DistanceMeters float64
}

type LocationPost struct {
	ID           string
	UserID       uint64
	ImageURL     string
	Caption      string
	LocationName string
	Latitude     float64
	Longitude    float64
}
