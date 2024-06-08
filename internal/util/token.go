package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	charSet     = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	checksumSet = "*~$=U"
)

func ReturnSignedToken(token string) string {
	checksum := getChecksum(calulateChecksum([]byte(token)))
	return fmt.Sprintf("%s%s", token, checksum)
}

func SignedToken(length int) string {
	if length <= 1 {
		return ""
	}
	bytes := generateBytes(length - 1)
	checksum := calulateChecksum(bytes)
	checksumChar := getChecksum(checksum)
	return fmt.Sprintf("%s%s", string(bytes), checksumChar)
}

func UnsignedToken(length int) string {
	if length <= 0 {
		return ""
	}
	return string(generateBytes(length))
}

func VerifyToken(signedToken string) bool {
	if len(signedToken) <= 1 {
		return false
	}
	submittedToken := signedToken[:len(signedToken)-1]
	submittedChecksum := signedToken[len(signedToken)-1]
	submittedIdent := []byte(submittedToken)

	checksum := calulateChecksum(submittedIdent)
	checksumChar := getChecksum(checksum)
	return string(submittedChecksum) == checksumChar
}

func calulateChecksum(bytes []byte) int {
	var rawSum int
	for _, i := range bytes {
		rawSum = rawSum + int(i)
	}
	return rawSum % 37
}

func getChecksum(checksumNumber int) string {
	checksumCharSet := strings.Split(fmt.Sprintf("%s%s", charSet, checksumSet), "")
	return checksumCharSet[checksumNumber]
}

func generateBytes(length int) []byte {
	rand.NewSource(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return result
}
