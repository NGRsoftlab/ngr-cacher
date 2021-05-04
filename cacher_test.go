// Copyright 2020 NGR Softlab
//
package casher

import (
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	locCache := New(5*time.Minute, 10*time.Minute)
	locCache.Set("a", "aaa", 60*time.Minute)

	k, ok := locCache.Get("a")
	if !ok {
		t.Error("Bad TestSet1")
	} else {
		if k != "aaa" {
			t.Error("Expected not this")
		}
	}

	err := locCache.Delete("a")
	if err != nil {
		t.Error("Bad TestSet2")
	}
}
