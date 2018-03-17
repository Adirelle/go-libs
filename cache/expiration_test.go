package cache

import (
	"testing"
	"time"
)

type FakeClock time.Time

func (f *FakeClock) Now() time.Time { return time.Time(*f) }

func (f *FakeClock) Advance(d time.Duration) {
	*f = FakeClock(time.Time(*f).Add(d))
}
func TestExpiringCache(t *testing.T) {

	cl := FakeClock(time.Unix(0, 0))

	c := NewMemoryStorage(
		Spy(func(s string, a ...interface{}) {
			t.Logf("%ds: "+s, append([]interface{}{cl.Now().Unix()}, a...)...)
		}),
		ExpirationUsingClock(8*time.Second, &cl),
	)

	if err := c.Set(5, 6); err != nil {
		t.Fatal("Set: expected <nil>")
	}

	if v, err := c.Get(5); err != nil || v != 6 {
		t.Fatal("Get: expected 6, <nil>")
	}

	cl.Advance(5 * time.Second)

	if v, err := c.Get(5); err != nil || v != 6 {
		t.Error("Get: expected 6, <nil>")
	}

	if err := c.Set(7, 8); err != nil {
		t.Error("Set: expected <nil>")
	}

	cl.Advance(10 * time.Second)

	if v, err := c.Get(5); err != ErrKeyNotFound || v != nil {
		t.Errorf("Get: expected <nil>, %s", ErrKeyNotFound)
	}

	if v, err := c.Get(7); err != ErrKeyNotFound || v != nil {
		t.Errorf("Get: expected <nil>, %s", ErrKeyNotFound)
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}
