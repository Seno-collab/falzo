package query

type Session struct {
	SessionID            string
	UserID               uint64
	Username             string
	Subject              string
	RefreshTokenHash     string
	RefreshExpiresAtUnix int64
}
