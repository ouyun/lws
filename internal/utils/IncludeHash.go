package utils

import "bytes"

func IncludeHash(hash []byte, hashList [][]byte) bool {
	if hashList == nil {
		return false
	}
	for _, target := range hashList {
		if bytes.Compare(hash, target) == 0 {
			return true
		}
	}
	return false
}
