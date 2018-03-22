package cache

import "testing"

func TestEmiter(t *testing.T) {

	ch := make(chan Event, 1)

	c := NewVoidStorage(Emitter(ch), Spy(t.Logf))

	c.Get(5)
	if e := <-ch; e.Type != GET || e.Key != 5 || e.Value != nil || e.Err != ErrKeyNotFound {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Put(5, 6)
	if e := <-ch; e.Type != PUT || e.Key != 5 || e.Value != 6 || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Remove(5)
	if e := <-ch; e.Type != REMOVE || e.Key != 5 || e.Value != false || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Flush()
	if e := <-ch; e.Type != FLUSH || e.Key != nil || e.Value != nil || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Len()
	if e := <-ch; e.Type != LEN || e.Key != nil || e.Value != 0 || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}
}
