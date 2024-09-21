package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5Code(value string) string {
	hashedPassword := md5.New()
	hashedPassword.Write([]byte(value))
	return hex.EncodeToString(hashedPassword.Sum(nil))
}
