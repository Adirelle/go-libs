package cache

import (
	"fmt"
	"time"
)

func ExampleExpiringCache() {

	b := new(MemoryStorage)
	cl := NewFackClock()
	c := NewExpiringCacheWithClock(b, 8*time.Second, cl)

	fmt.Println(c.Set(5, 6))
	fmt.Println(c.Get(5))

	cl.Advance(5 * time.Second)

	fmt.Println(c.Get(5))
	fmt.Println(c.Set(7, 8))
	fmt.Println(c.Get(7))

	cl.Advance(10 * time.Second)

	fmt.Println(c.Get(5))
	fmt.Println(c.Get(7))

	// Output:
	// <nil>
	// 6 <nil>
	// 6 <nil>
	// <nil>
	// 8 <nil>
	// <nil> Key not found
	// <nil> Key not found
}
