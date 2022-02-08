package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

func Sha256(passwd string) string {
	rh := sha256.New()
	rh.Write([]byte(passwd))
	return hex.EncodeToString(rh.Sum(nil))
}
func GetToken(src string) string {
	h := md5.New()
	h.Write([]byte(src + strconv.FormatInt(time.Now().Unix(), 10)))
	return hex.EncodeToString(h.Sum(nil))
}
