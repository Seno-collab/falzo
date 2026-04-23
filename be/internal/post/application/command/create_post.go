package command

type CreatePost struct {
	UserID       uint64
	ImageURL     string
	Caption      string
	LocationName string
	Latitude     float64
	Longitude    float64
}
