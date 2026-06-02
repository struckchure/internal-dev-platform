package internals

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type IHasher interface {
	HashPassword(password string) string
	PasswordIsCorrect(hashedPassword string, password string) bool
}

type Hasher struct{}

const (
	argon2Time    uint32 = 1
	argon2Memory  uint32 = 64 * 1024
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	saltLen              = 16
)

func parseArgon2Hash(hash string) (memory uint32, time uint32, threads uint8, salt, expected []byte, ok bool) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return 0, 0, 0, nil, nil, false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, false
	}
	expected, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, false
	}
	return memory, time, threads, salt, expected, true
}

func (*Hasher) HashPassword(password string) string {
	if len(password) == 0 {
		return ""
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return ""
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func (*Hasher) PasswordIsCorrect(hashedPassword string, password string) bool {
	if strings.HasPrefix(hashedPassword, "$2a$") || strings.HasPrefix(hashedPassword, "$2b$") || strings.HasPrefix(hashedPassword, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
	}

	memory, time, threads, salt, expected, ok := parseArgon2Hash(hashedPassword)
	if !ok {
		return false
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expected)))

	return subtle.ConstantTimeCompare(hash, expected) == 1
}

func NewHasher() IHasher {
	return &Hasher{}
}
