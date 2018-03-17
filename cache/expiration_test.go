package cache

import (
	"fmt"
	"time"
)

func ExampleExpiringCache() {

	b := NewMemoryStorage()
	cl := NewFakeClock()
	c := New(b, ExpirationUsingClock(8*time.Second, cl))

	fmt.Println(c.Set(5, 6))
	fmt.Println(c.Get(5))

	cl.Advance(5 * time.Second)

	fmt.Println(c.Get(5))
	fmt.Println(c.Set(7, 8))
	fmt.Println(c.Get(7))

	cl.Advance(10 * time.Second)

	fmt.Println(c.Get(5))
	fmt.Println(c.Get(7))
}
