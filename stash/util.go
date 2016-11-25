package stash

import (
	"encoding/base64"
	"hash"
)

func ComputeChecksum(name string, hasher hash.Hash) []byte {
	hasher.Write([]byte(name))
	return hasher.Sum(nil)
}

func EncodeChecksum(p []byte, encoding *base64.Encoding) string {
	return encoding.EncodeToString(p)
}
