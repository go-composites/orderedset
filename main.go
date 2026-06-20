package main

import (
	"fmt"

	OrderedSet "github.com/go-composites/orderedset/src"
	Result "github.com/go-composites/result/src"
)

func main() {
	a := OrderedSet.New(1, 2, 3)
	a.Add(3).Add(4) // Add returns the set, so calls chain; 3 is idempotent.

	fmt.Printf("Len = %d\n", a.Len())
	fmt.Printf("Has(2) = %t\n", a.Has(2))
	fmt.Printf("IsEmpty = %t\n", a.IsEmpty())
	fmt.Printf("order = %v\n", order(a)) // [1 2 3 4] — deterministic.

	b := OrderedSet.New(3, 4, 5)
	fmt.Printf("Union        = %v\n", order(a.Union(b)))
	fmt.Printf("Intersection = %v\n", order(a.Intersection(b)))
	fmt.Printf("Difference   = %v\n", order(a.Difference(b)))
	fmt.Printf("IsSubset     = %t\n", OrderedSet.New(1, 2).IsSubset(a))
	fmt.Printf("Equal        = %t\n", a.Equal(OrderedSet.New(4, 3, 2, 1)))

	a.Delete(1)
	fmt.Printf("after Delete(1), order = %v\n", order(a))
}

// order collects an OrderedSet's int items in iteration order — which, unlike a
// plain Set, is deterministic (first-insertion order), so no sorting is needed.
func order(s OrderedSet.Interface) []int {
	out := []int{}
	s.Each(func(item interface{}) Result.Interface {
		out = append(out, item.(int))
		return Result.New()
	})
	return out
}
