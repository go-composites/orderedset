<p align="center"><img src="https://raw.githubusercontent.com/go-composites/brand/main/social/go-composites.png" alt="go-composites/orderedset" width="720"></p>

# orderedset

[![ci](https://github.com/go-composites/orderedset/actions/workflows/ci.yml/badge.svg)](https://github.com/go-composites/orderedset/actions/workflows/ci.yml)

An **insertion-ordered** Set composite for Composition-Oriented Programming — a
collection of unique items that remembers the order in which each item was first
added. It is the deterministic sibling of [`set`](https://github.com/go-composites/set)
(unordered) and shares the `Each`/`ToArray` grammar with
[`array`](https://github.com/go-composites/array).

An `OrderedSet` is backed by **both** a `[]interface{}` (recording
first-insertion order) and a `map[interface{}]struct{}` (giving O(1) membership
and dedup), kept in lock-step. The result: iteration, `ToArray` and the
set-algebra results are all emitted in a **deterministic, first-insertion
order** rather than Go's unspecified map order — so consumers (and tests) never
depend on map-iteration flakiness.

It follows the go-composites grammar:

- **Deterministic order**: `Add` appends only when an item is new (a duplicate
  keeps its original position); `Delete` removes while preserving the order of
  the remaining items; `Each`/`ToArray` iterate in insertion order; `Union`,
  `Intersection` and `Difference` all yield a deterministic order (the
  receiver's order, then — for `Union` — the other's new items in its order).
- **Never nil / Null-Object**: every constructor and method returns a real
  object; `Null()` provides an inert variant and `IsNull()` distinguishes it.
- **Result-based errors**: fallible iteration returns a
  [`Result`](https://github.com/go-composites/result) — `Each` short-circuits on
  the first `Result` whose `HasError()` is true. No panics, no bare nils.
- **Composite returns**: `ToArray()` materialises into an
  [`Array`](https://github.com/go-composites/array); set algebra returns fresh
  `OrderedSet`s.

Items must be comparable, since an `OrderedSet` is backed by a Go map for
membership.

**`Equal` is order-INsensitive**: two OrderedSets are equal when they hold the
same members regardless of insertion order, exactly like a mathematical set.

## Install

```sh
go get github.com/go-composites/orderedset@main
```

## Usage

```go
package main

import (
	"fmt"

	OrderedSet "github.com/go-composites/orderedset/src"
	Result "github.com/go-composites/result/src"
)

func main() {
	a := OrderedSet.New(1, 2, 3)
	a.Add(3).Add(4) // Add returns the set, so calls chain; 3 is idempotent.

	fmt.Println(a.Len())     // 4
	fmt.Println(a.Has(2))    // true
	fmt.Println(a.IsEmpty()) // false

	b := OrderedSet.New(3, 4, 5)
	_ = a.Union(b)        // {1,2,3,4,5} — receiver order, then other's new items
	_ = a.Intersection(b) // {3,4}       — receiver order
	_ = a.Difference(b)   // {1,2}       — receiver order

	fmt.Println(OrderedSet.New(1, 2).IsSubset(a))    // true
	fmt.Println(a.Equal(OrderedSet.New(4, 3, 2, 1))) // true (order-insensitive)

	// Each iterates in insertion order and short-circuits on the first
	// Result whose HasError() is true.
	a.Each(func(item interface{}) Result.Interface {
		fmt.Println(item)
		return Result.New()
	})

	// ToArray materialises into a go-composites Array, in insertion order.
	_ = a.ToArray()

	a.Delete(1) // removes 1, keeping {2,3,4} in order
}
```

### API

| Method | Returns | Notes |
| --- | --- | --- |
| `New(items...)` | `OrderedSet.Interface` | variadic, deduplicated, first-seen order |
| `Null()` | `OrderedSet.Interface` | inert Null-Object; `IsNull()` is `true` |
| `Add(item)` | `OrderedSet.Interface` | appends only if new (keeps original position); chainable |
| `Delete(item)` | `OrderedSet.Interface` | no-op when absent; preserves remaining order; chainable |
| `Has(item)` | `bool` | membership test (O(1)) |
| `Len()` | `int` | number of items |
| `IsEmpty()` | `bool` | `true` when there are no items |
| `Each(fn)` | `Result.Interface` | iterate in insertion order; short-circuit on `HasError()` |
| `ToArray()` | `Array.Interface` | materialise into an Array, in insertion order |
| `Union(other)` | `OrderedSet.Interface` | Ruby `\|` — receiver order, then other's new items |
| `Intersection(other)` | `OrderedSet.Interface` | Ruby `&` — common items, receiver order |
| `Difference(other)` | `OrderedSet.Interface` | Ruby `-` — receiver items not in other, receiver order |
| `IsSubset(other)` | `bool` | every item is also in `other` |
| `Equal(other)` | `bool` | same members, **order-insensitive** |
| `IsNull()` | `bool` | `false` for a real OrderedSet |

## License

BSD-3-Clause — see [LICENSE](LICENSE).
