package mqtt

import (
	"testing"
)

func TestNewRedisPool(t *testing.T) {
	pool := NewRedisPool()
	red := pool.Get()
	err := red.Err()
	if err != nil {
		t.Error("conn redis fail")
	}
	defer red.Close()
}
