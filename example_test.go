package cache_test

import (
	"fmt"
	lrucache "lrucache"
)

// This example demonstrates basic usage of the LRU cache and is used
// by `go doc` and `go test` as documentation.
func Example() {
	c := lrucache.NewLRU(2)
	c.Add("a", "alpha")
	c.Add("b", "beta")

	if v, ok := c.Get("a"); ok {
		fmt.Println(v)
	}

	// adding a third element evicts the least-recently-used one
	c.Add("c", "gamma")

	if _, ok := c.Get("b"); !ok {
		fmt.Println("b evicted")
	}

	// Output:
	// alpha
	// b evicted
}
