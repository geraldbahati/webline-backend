package mpesa

import (
	"encoding/base64"
)

func GeneratePassword(businessShortCode, passkey, timestamp string) string {
    password := businessShortCode + passkey + timestamp
    return base64.StdEncoding.EncodeToString([]byte(password))
}
