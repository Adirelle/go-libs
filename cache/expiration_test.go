package cache

import (
	"fmt"
	"time"
)

type FakeClock time.Time

func (f *FakeClock) Now() time.Time { return time.Time(*f) }

func (f *FakeClock) Advance(d time.Duration) {
	*f = FakeClock(time.Time(*f).Add(d))
}
func ExampleExpiringCache() {

	cl := FakeClock(time.Unix(1, 1))
	c := NewMemoryStorage(ExpirationUsingClock(8*time.Second, &cl))

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
