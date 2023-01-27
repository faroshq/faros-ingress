package utilhash

import (
	"crypto/sha256"
	"strings"

	"github.com/martinlindhe/base36"
)

func GetHash(str string) string {
	hash := sha256.Sum224([]byte(str))
	return strings.ToLower(base36.EncodeBytes(hash[:]))
}
