package mongodb

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

// hash generates the hashed string by an input string.
func hash(s string) string {
	sum := sha1.Sum([]byte(s))
	return strings.ToUpper(fmt.Sprintf("%x", sum))
}
