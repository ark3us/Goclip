package common

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

const AppName = "Goclip"

func Md5Digest(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func TimeToString(t time.Time) string {
	// return t.Format("2006-01-02 15:04:05")
	return t.Format("15:04:05")
}
