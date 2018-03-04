package dic

import (
	"fmt"
	"strconv"
)

func ExampleConstant() {
	// Container setup
	ctn := New()
	ctn.Register(Constant("/etc/hosts"))

	// Container use
	var path string
	if err := ctn.Fetch(&path); err != nil {
		panic(err)
	}
	fmt.Print(path)
	// Output:
	// /etc/hosts
}

func ExampleFunc() {
	// Container setup
	ctn := New()
	ctn.Register(Func(strconv.Itoa))
	ctn.Register(Constant(25))

	// Container use
	var s string
	if err := ctn.Fetch(&s); err != nil {
		panic(err)
	}

	fmt.Println(s)
	// Output:
	// 25
}

func ExampleSingleton() {
	// Container setup
	ctn := New()
	// Func returns an already-Singleton-wrapped provider
	ctn.Register(Func(func() int {
		fmt.Println("Called !")
		return 5
	}))

	// Container use
	var a, b int
	if err := ctn.Fetch(&a); err != nil {
		panic(err)
	}
	if err := ctn.Fetch(&b); err != nil {
		panic(err)
	}
	fmt.Println(a, b)
	// Output:
	// Called !
	// 5 5
}

func ExampleCycleError() {
	// Container setup
	ctn := New()
	ctn.Register(Func(strconv.Itoa)) // func Itoa(i int) string
	ctn.Register(Func(strconv.Atoi)) // func Atoi(s string) (int, error)

	// Container use
	var a int
	err := ctn.Fetch(&a)
	fmt.Print(err)
	// Output:
	// cannot inject argument #0 of func(string) (int, error):
	// 	cannot inject argument #0 of func(int) string:
	// 	cycle involving these providers: [Singleton(func(string) (int, error)) Singleton(func(int) string)]
}
