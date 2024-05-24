// Copyright 2020-2024 (c) NGR Softlab

package casher

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestPlainSetGetDeleteScenario(t *testing.T) {
	locCache := New(5*time.Minute, 10*time.Minute)

	// no need refresh
	locCache.Set("a", "aaa", 10*time.Second)
	time.Sleep(9 * time.Second)

	k, ok := locCache.Get("a")
	if !ok {
		t.Error("Bad TestSet1")
	} else {
		if k != "aaa" {
			t.Error("Expected not this")
		}
	}

	// no need refresh
	locCache.Set("b", "bbb", 10*time.Second, ItemOptions{
		NeedRefresh: false,
	})
	time.Sleep(9 * time.Second)

	k, ok = locCache.Get("b")
	if !ok {
		t.Error("Bad TestSet2")
	} else {
		if k != "bbb" {
			t.Error("Expected not this")
		}
	}

	// wait more time (> 10*time.Second after set operation) to check expiration refresh
	time.Sleep(6 * time.Second)
	k, ok = locCache.Get("b")
	if ok {
		t.Error("Bad TestSet3")
	}
	t.Log("val ", k)

	// need refresh
	locCache.Set("c", "ccc", 10*time.Second, ItemOptions{
		NeedRefresh: true,
	})
	time.Sleep(9 * time.Second)

	k, ok = locCache.Get("c")
	if !ok {
		t.Error("Bad TestSet4")
	} else {
		if k != "ccc" {
			t.Error("Expected not this")
		}
	}

	// wait more time (> 10*time.Second after set operation) to check expiration refresh
	time.Sleep(6 * time.Second)
	k, ok = locCache.Get("c")
	if !ok {
		t.Error("Bad TestSet5")
	}
	t.Log("val ", k)

	///////////////////////////
	// clean data
	err := locCache.Delete("a")
	if err != nil {
		t.Error("Bad TestSet6")
	}
	err = locCache.Delete("b")
	if err != nil {
		t.Error("Bad TestSet7")
	}
	err = locCache.Delete("c")
	if err != nil {
		t.Error("Bad TestSet8")
	}
}

func TestCacheScenario(t *testing.T) {
	type ItemTest struct {
		item        interface{}
		onCloseF    func(item interface{}) error
		NeedOnClose bool
		NeedRefresh bool
	}

	tests := []struct {
		name         string
		okResTimeout time.Duration
		noResTimeout time.Duration
		checkData    map[string]ItemTest
	}{
		{
			name:         "valid",
			okResTimeout: 1 * time.Second,
			noResTimeout: 6 * time.Second,
			checkData: map[string]ItemTest{
				"test1": {
					item: 1,
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						return nil
					},
					NeedOnClose: true,
				},
				"test2": {
					item: "test2 item",
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						return nil
					},
					NeedOnClose: true,
				},
				"test3": {
					item: make(chan int),
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						itemCast, ok := item.(chan int)
						if !ok {
							return errors.New("bad type cast")
						}
						close(itemCast)

						return nil
					},
					NeedOnClose: true,
				},
				"test5": {
					item:        2,
					NeedOnClose: false,
				},
			},
		},
		{
			name:         "valid (nil db close panic processing check)",
			okResTimeout: 1 * time.Second,
			noResTimeout: 6 * time.Second,
			checkData: map[string]ItemTest{
				"test1": {
					item: &sql.DB{},
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						itemCast, ok := item.(*sql.DB)
						if !ok {
							return errors.New("bad type cast")
						}

						err := itemCast.Close()
						if err != nil {
							return errors.New("bad item close")
						}

						return nil
					},
					NeedOnClose: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LocalCache := New(10*time.Second, 5*time.Second)
			defer LocalCache.ClearAll()

			for k, v := range tt.checkData {
				if v.NeedOnClose {
					LocalCache.Set(k, v.item, 5*time.Second, ItemOptions{
						OnDeleteFunc: v.onCloseF,
						NeedOnDelete: true,
					})
				} else {
					LocalCache.Set(k, v.item, 5*time.Second)
				}
			}

			time.Sleep(tt.okResTimeout)
			for k := range tt.checkData {
				_, ok := LocalCache.Get(k)
				if !ok {
					t.Error("no cached result found")
				}
			}

			time.Sleep(tt.noResTimeout)
			for k := range tt.checkData {
				_, ok := LocalCache.Get(k)
				if ok {
					t.Error("cached result was not deleted after timeout")
				}
			}
		})
	}
}
