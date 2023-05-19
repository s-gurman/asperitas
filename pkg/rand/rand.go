package rand

import (
	"crypto/rand"
	"fmt"
)

const LengthOfID = 24

func GetRandID() string {
	randID := make([]byte, LengthOfID/2)
	rand.Read(randID) //nolint:errcheck
	return fmt.Sprintf("%x", randID)
}
