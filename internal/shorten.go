package internal

import (
	"math/rand"
	"time"
)

// Use k-v store might be better than using Postgres (we only need hash : real_url)

// But there is a way we can do it with Postgres, or even sqlite, that we use auto increment ID => encode into string

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// Rate of collision: Extremely low IMO, 2 calls exactly the same nanosecond to potentially generate the same sequence
// Thanks to high resolution of UnixNano

// Discussion
// How about using encode from auto-increment id from database ? (Already think about it, what if the url is stale ? Remove it and can't do anything with that id again ?)
// Any idea ???
