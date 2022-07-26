package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

// RememberToken is a helper to generate remeber tiokens of a predetermined byte size
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}

/*
Bytes is used to generate n random bytes or return an error
Makes use of crypto/rand
*/
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String is used to generate a byte of size nBytes and return a string that is base64 URL encoded
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
