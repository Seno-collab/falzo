package security

import (
	"crypto/rand"
	"fmt"
)

const inviteCodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

type InviteCodeGenerator struct {
	length int
}

func NewInviteCodeGenerator(length int) *InviteCodeGenerator {
	return &InviteCodeGenerator{length: length}
}

func (g *InviteCodeGenerator) Generate() (string, error) {
	if g.length <= 0 {
		return "", fmt.Errorf("invite code length must be positive")
	}

	buffer := make([]byte, g.length)
	random := make([]byte, g.length)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate invite code: %w", err)
	}
	for index, value := range random {
		buffer[index] = inviteCodeAlphabet[int(value)%len(inviteCodeAlphabet)]
	}
	return string(buffer), nil
}
