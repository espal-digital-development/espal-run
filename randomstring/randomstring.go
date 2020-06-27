package randomstring

import (
	"math/rand"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type RandomString struct {
	src rand.Source
}

// Simple returns a simple random string in the requested length.
func (r *RandomString) Simple(n int) string {
	buffer := make([]byte, n)
	// A randomStringSrc.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, r.src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			buffer[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(buffer)
}

// New returns a new instance of RandomString.
func New() (*RandomString, error) {
	r := &RandomString{
		src: rand.NewSource(time.Now().UnixNano()),
	}
	return r, nil
}
