package models

import (
	"os"
)

var SecretKey = os.Getenv("SECRET_KEY")

type User struct {
	ID             int
	Guid           string
	RefreshToken   string
	RefreshTokenID string
	Used           bool
	IpAddress      string
}
