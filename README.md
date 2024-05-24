# ngr-cacher
Lib with little in-memory cacher (key - string, value - interface{}).
In last version onDelete item function can be specified (experimental) 

# import
```
import "github.com/NGRsoftlab/ngr-cacher"
```

# params
```
Cache:
    sync.RWMutex - 
    items - map[string]interface{}, items map
    defaultExpiration - default expiration time to delete items, which exceeded it
    cleanupInterval - GC iteration timeout
    
Item:
    Value - interface{}
	Expiration - int64, expiration time of item
	Created - time.Time, item creation time
	Options - ItemOptions, experimental, item specific params

ItemOptions:
    NeedOnDelete - bool, on/off using on-delete func when item's deleted from cache
	OnDeleteFunc - func(item interface{}) error, on-delete func
	NeedRefresh - bool, on/off expiration time update when item't got from cache
	refreshDuration - time.Duration, expiration item timeout backup for refreshing
```

# plain example
```
// create cache instance with default expiration time = 5min, GC iteration timeout = 10min
locCache := New(5*time.Minute, 10*time.Minute)

// set string value into "testKey" element, set element expiration time = 60min
// if you prefer using default expiration - just set 0 into this param
locCache.Set("testKey", "testValue", 60*time.Minute)

// get
val, ok := locCache.Get("testKey")
	if !ok {
		fmt.Println("no such key")
	}
	
// delete (no such key in cache)
err := locCache.Delete("a?")
	if err != nil {
		fmt.Printf("can't delete: %s", err.Error())
	}
```

# with refresh example
```
// create cache instance with default expiration time = 5min, GC iteration timeout = 10min
locCache := New(5*time.Minute, 10*time.Minute)

// set string value into "testKey" element, set element expiration time = 60min
// switch on refreshing item expiration time when using get-item on cache instance
// if you prefer using default expiration - just set 0 into this param
locCache.Set("testKey", "testValue", 60*time.Minute, ItemOptions{
		NeedRefresh: true,
	})

// get
val, ok := locCache.Get("testKey")
	if !ok {
		fmt.Println("no such key")
	}
	
// delete (no such key in cache)
err := locCache.Delete("a?")
	if err != nil {
		fmt.Printf("can't delete: %s", err.Error())
	}
```

# with onCloseFunc example (experimental feature)
```
// create cache instance with default expiration time = 5min, GC iteration timeout = 10min
locCache := New(5*time.Minute, 10*time.Minute)
onCloseFunc := func(item interface{}) error {
				itemCast, ok := reflect.ValueOf(item).Interface().(chan int)
				if !ok {
					return errors.New("bad type cast")
				}
				close(itemCast)

				return nil
			}
					
// set chan int value into "testChanKey" element, set element expiration time = 60min, set onCloseFunc
// if you prefer using default expiration - just set 0 into this param
locCache.Set("testChanKey", make(chan int), 60*time.Minute, 
            ItemOptions{OnDeleteFunc: onCloseFunc, NeedOnDelete: true,})

// get
val, ok := locCache.Get("testChanKey")
	if !ok {
		fmt.Println("no such key")
	}

// delete (chan close func will be called inside locCache.Delete)
err := locCache.Delete("testChanKey")
	if err != nil {
		fmt.Printf("can't delete: %s", err.Error())
	}
```

# tests
```
just run go test
```
