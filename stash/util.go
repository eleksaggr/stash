package stash

import (
	"encoding/base64"
	"hash"
)

func computeChecksum(name string, hasher hash.Hash) []byte {
	hasher.Write([]byte(name))
	return hasher.Sum(nil)
}

func encodeChecksum(p []byte, encoding *base64.Encoding) string {
	return encoding.EncodeToString(p)
}
