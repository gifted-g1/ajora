package utils

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "strings"

    "github.com/google/uuid"
)

func NewReference(prefix string) string {
    b := make([]byte, 8)
    _, _ = rand.Read(b)
    return fmt.Sprintf("%s_%s_%s", prefix, strings.ToUpper(hex.EncodeToString(b[:4])), uuid.New().String()[:8])
}

func GenerateInviteCode() string {
    b := make([]byte, 4)
    _, _ = rand.Read(b)
    return fmt.Sprintf("AJ-%s", strings.ToUpper(hex.EncodeToString(b)))
}

func GenerateOTP() string {
    b := make([]byte, 4)
    _, _ = rand.Read(b)
    n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
    return fmt.Sprintf("%06d", n)
}
