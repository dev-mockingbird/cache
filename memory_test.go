package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	m := MemoryCache("test")
	ctx := context.Background()
	var total int
	for i := 0; i < 100000; i++ {
		total += i
		if err := m.Put(ctx, fmt.Sprintf("test-%d", i), i); err != nil {
			t.Fatal(err)
		}
	}
	var r int
	for i := 0; i < 100000; i++ {
		var v int
		if err := m.Get(ctx, fmt.Sprintf("test-%d", i), &v); err != nil {
			t.Fatal(err)
		}
		r += v
	}
	if total != r {
		t.Fatal("not equal")
	}
	for i := 0; i < 100000; i++ {
		if err := m.Del(ctx, fmt.Sprintf("test-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.(*memoryCache).data) != 0 {
		t.Fatal("del error")
	}
	if err := m.Put(ctx, "test", 1, time.Microsecond*500); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	var v int
	if err := m.Get(ctx, "test", &v); err == nil || !errors.Is(err, ErrCacheNotFound) {
		t.Fatal("timeout error")
	}
}
