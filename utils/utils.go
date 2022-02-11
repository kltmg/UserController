package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"path"
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

// GetFileName 为上传的文件生成一个文件名.
func GetFileName(fileName string, ext string) string {
	h := md5.New()
	h.Write([]byte(fileName + strconv.FormatInt(time.Now().Unix(), 10)))
	return hex.EncodeToString(h.Sum(nil)) + ext
}

// CheckAndCreateFileName 检查文件后缀合法性.
func CheckAndCreateFileName(oldName string) (newName string, isLegal bool) {
	ext := path.Ext(oldName)
	isLegal = false
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
		//随机生成一个文件名.
		newName = GetFileName(oldName, ext)
		isLegal = true
	}
	return newName, isLegal
}
