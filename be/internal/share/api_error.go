type share

type ApiError struct {
	status  int
	message string
	code    string
	field   string
	detail  string
	logErr  bool
}