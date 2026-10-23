
package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	cache := NewCache(5 * time.Second)

	cache.Add("test-url", []byte("hello"))

	val, ok := cache.Get("test-url")
	if !ok {
		t.Fatal("expected to find key")
	}

	if string(val) != "hello" {
		t.Errorf("expected hello, got %s", string(val))
	}
}
