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

func TestCacheScenario(t *testing.T) {
	type ItemTest struct {
		item        interface{}
		onCloseF    func(item interface{}) error
		itemType    reflect.Type
		needOptions bool
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
					itemType:    reflect.TypeOf(reflect.Int),
					needOptions: true,
				},
				"test2": {
					item: "test2 item",
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						return nil
					},
					itemType:    reflect.TypeOf(reflect.String),
					needOptions: true,
				},
				"test3": {
					item: make(chan int),
					onCloseF: func(item interface{}) error {
						fmt.Printf("Close.item: %v, itemType: %v\n", item, reflect.TypeOf(item))
						itemCast, ok := reflect.ValueOf(item).Interface().(chan int)
						if !ok {
							return errors.New("bad type cast")
						}
						close(itemCast)

						return nil
					},
					itemType:    reflect.TypeOf(reflect.Chan),
					needOptions: true,
				},
				"test5": {
					item:        2,
					needOptions: false,
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
						itemCast, ok := reflect.ValueOf(item).Interface().(*sql.DB)
						if !ok {
							return errors.New("bad type cast")
						}

						err := itemCast.Close()
						if err != nil {
							return errors.New("bad item close")
						}

						return nil
					},
					itemType:    reflect.TypeOf(reflect.Chan),
					needOptions: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LocalCache := New(10*time.Second, 5*time.Second)
			defer LocalCache.ClearAll()

			for k, v := range tt.checkData {
				if v.needOptions {
					LocalCache.Set(k, v.item, 5*time.Second, ItemOptions{
						onDeleteFunc: v.onCloseF,
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
