package goclip

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

const AppName = "Goclip"
const AppId = "net.ark3us.goclip"

func Md5Digest(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func TimeToString(t time.Time, full bool) string {
	if full {
		return t.Format("2006-01-02 15:04:05")
	} else {
		return t.Format("15:04:05")
	}
}
