package cache

import (
	"fmt"
	"sync"
	"time"
)

func ExampleVoidStorage() {

	n := VoidStorage{}

	fmt.Println(n.Get(5))
	fmt.Println(n.Set(5, 6))
	fmt.Println(n.Get(5))
	fmt.Println(n.GetIFPresent(5))
	fmt.Println(n.Remove(5))
	fmt.Println(n.Get(5))

	// Output:
	// <nil> Key not found
	// <nil>
	// <nil> Key not found
	// <nil> Key not found
	// false
	// <nil> Key not found
}

func ExampleMemoryStorage() {

	c := NewMemoryStorage()

	fmt.Println(c.Get(5))
	fmt.Println(c.Set(5, 6))
	fmt.Println(c.Get(5))
	fmt.Println(c.GetIFPresent(5))
	fmt.Println(c.Remove(5))
	fmt.Println(c.Remove(5))
	fmt.Println(c.Get(5))

	// Output:
	// <nil> Key not found
	// <nil>
	// 6 <nil>
	// 6 <nil>
	// true
	// false
	// <nil> Key not found
}

func ExampleLoader() {

	c := Loader(func(k interface{}) (interface{}, error) {
		fmt.Println("Load", k)
		return 6, nil
	})

	fmt.Println(c.Get(5))
	fmt.Println(c.Set(5, 6))
	fmt.Println(c.Get(5))
	fmt.Println(c.GetIFPresent(5))
	fmt.Println(c.Remove(5))
	fmt.Println(c.Remove(5))
	fmt.Println(c.Get(5))

	// Output:
	// Load 5
	// 6 <nil>
	// <nil>
	// Load 5
	// 6 <nil>
	// <nil> Key not found
	// false
	// false
	// Load 5
	// 6 <nil>
}

func ExampleLockingCache() {

	l := Loader(func(k interface{}) (interface{}, error) {
		time.Sleep(time.Duration(k.(int)) * time.Millisecond)
		return k, nil
	})

	c := New(l, Locking)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Get(500)")
	go func() {
		defer wg.Done()
		fmt.Println(c.Get(500))
	}()

	fmt.Println("Get(200)")
	go func() {
		defer wg.Done()
		fmt.Println(c.Get(200))
	}()

	wg.Wait()

	// Output:
	// Get(500)
	// Get(200)
	// 200 <nil>
	// 500 <nil>
}
