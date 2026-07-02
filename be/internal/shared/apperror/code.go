package apperror

type Code string

const (
	CodeInvalidRequest      Code = "INVALID_REQUEST"
	CodeValidationError     Code = "VALIDATION_ERROR"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeInternalServerError Code = "INTERNAL_SERVER_ERROR"

	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
	CodeUserLocked         Code = "USER_LOCKED"
	CodeUserDisabled       Code = "USER_DISABLED"
	CodeTokenExpired       Code = "TOKEN_EXPIRED"

	CodeWalletNotEnoughBalance Code = "WALLET_NOT_ENOUGH_BALANCE"
	CodeMatchNotFound          Code = "MATCH_NOT_FOUND"
	CodeGameRoomFull           Code = "GAME_ROOM_FULL"
)
