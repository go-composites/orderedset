package OrderedSet

import (
	Array "github.com/go-composites/array/src"
	Result "github.com/go-composites/result/src"
)

// Interface is the public contract of an OrderedSet composite — a collection of
// unique items that remembers the order in which each item was first added. It
// is the insertion-ordered sibling of Set (and shares the Each/ToArray grammar
// with Array): iteration, ToArray and the set-algebra results are all emitted in
// a deterministic, first-insertion order rather than Go's unspecified map order.
//
// Items are arbitrary comparable values (the OrderedSet is backed by a Go map
// for O(1) membership, so each item must be comparable, exactly like a
// Dictionary key). Membership tests return a plain bool, fallible iteration
// returns a Result, and every method honours the Null-Object invariant (never
// nil).
type Interface interface {
	Add(item interface{}) Interface
	Delete(item interface{}) Interface
	Has(item interface{}) bool
	Len() int
	IsEmpty() bool
	Each(fn func(item interface{}) Result.Interface) Result.Interface
	ToArray() Array.Interface
	Union(other Interface) Interface
	Intersection(other Interface) Interface
	Difference(other Interface) Interface
	IsSubset(other Interface) bool
	Equal(other Interface) bool
	IsNull() bool
}

// data backs the OrderedSet with BOTH a slice (recording first-insertion order)
// and a map (giving O(1) membership and dedup). The two are kept in lock-step:
// an item lives in order iff it is a key of member.
type data struct {
	order  []interface{}
	member map[interface{}]struct{}
}

// New creates an OrderedSet seeded with the given items, deduplicated and kept
// in first-seen order. Items must be comparable, since the OrderedSet is backed
// by a Go map for membership.
func New(items ...interface{}) Interface {
	d := &data{
		order:  []interface{}{},
		member: make(map[interface{}]struct{}),
	}
	for _, item := range items {
		d.Add(item)
	}
	return d
}

// Add inserts item, appending it to the insertion order only when it is new (a
// no-op for an already-present item, which preserves its original position), and
// returns the receiver so calls chain.
func (d *data) Add(item interface{}) Interface {
	if _, ok := d.member[item]; !ok {
		d.member[item] = struct{}{}
		d.order = append(d.order, item)
	}
	return d
}

// Delete removes item from both the slice and the map (a no-op when absent),
// preserving the relative order of the remaining items, and returns the receiver
// so calls chain.
func (d *data) Delete(item interface{}) Interface {
	if _, ok := d.member[item]; !ok {
		return d
	}
	delete(d.member, item)
	for i, existing := range d.order {
		if existing == item {
			d.order = append(d.order[:i], d.order[i+1:]...)
			break
		}
	}
	return d
}

// Has reports whether item is a member of the OrderedSet.
func (d *data) Has(item interface{}) bool {
	_, ok := d.member[item]
	return ok
}

// Len returns the number of items in the OrderedSet.
func (d *data) Len() int {
	return len(d.order)
}

// IsEmpty reports whether the OrderedSet has no items.
func (d *data) IsEmpty() bool {
	return len(d.order) == 0
}

// Each iterates over the items in insertion order, invoking fn for each. It
// short-circuits and returns the first Result for which HasError() is true; on a
// full pass it returns a fresh Result.New().
func (d *data) Each(
	fn func(item interface{}) Result.Interface,
) Result.Interface {
	for _, item := range d.order {
		if result := fn(item); result.HasError() {
			return result
		}
	}
	return Result.New()
}

// ToArray materialises the OrderedSet into a go-composites Array, in insertion
// order.
func (d *data) ToArray() Array.Interface {
	arr := Array.New()
	for _, item := range d.order {
		arr.Push(item)
	}
	return arr
}

// Union returns a new OrderedSet containing every item present in this set or in
// other (Ruby's `|`). The receiver's items come first, in their order, followed
// by other's items not already present, in other's order.
func (d *data) Union(other Interface) Interface {
	result := New()
	for _, item := range d.order {
		result.Add(item)
	}
	other.Each(func(item interface{}) Result.Interface {
		result.Add(item)
		return Result.New()
	})
	return result
}

// Intersection returns a new OrderedSet containing only the items present in
// both this set and other (Ruby's `&`), in the receiver's order.
func (d *data) Intersection(other Interface) Interface {
	result := New()
	for _, item := range d.order {
		if other.Has(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference returns a new OrderedSet containing the items present in this set
// but not in other (Ruby's `-`), in the receiver's order.
func (d *data) Difference(other Interface) Interface {
	result := New()
	for _, item := range d.order {
		if !other.Has(item) {
			result.Add(item)
		}
	}
	return result
}

// IsSubset reports whether every item of this set is also in other.
func (d *data) IsSubset(other Interface) bool {
	for _, item := range d.order {
		if !other.Has(item) {
			return false
		}
	}
	return true
}

// Equal reports whether this set and other contain exactly the same members.
// Equality is order-INsensitive: two OrderedSets are equal when they hold the
// same items regardless of insertion order, exactly like a mathematical set.
func (d *data) Equal(other Interface) bool {
	return d.Len() == other.Len() && d.IsSubset(other)
}

// IsNull reports that this is a real (non-null) OrderedSet.
func (d *data) IsNull() bool {
	return false
}

// null is the Null-Object variant of an OrderedSet: an empty, immutable
// placeholder that honours the full Interface without ever being nil. Mutating
// methods are no-ops that return the receiver; queries are empty/false/zero.
type null struct{}

// Null returns the Null-Object OrderedSet.
func Null() Interface {
	return &null{}
}

func (n *null) Add(item interface{}) Interface { return n }

func (n *null) Delete(item interface{}) Interface { return n }

func (n *null) Has(item interface{}) bool { return false }

func (n *null) Len() int { return 0 }

func (n *null) IsEmpty() bool { return true }

func (n *null) Each(
	fn func(item interface{}) Result.Interface,
) Result.Interface {
	return Result.New()
}

func (n *null) ToArray() Array.Interface { return Array.New() }

func (n *null) Union(other Interface) Interface { return n }

func (n *null) Intersection(other Interface) Interface { return n }

func (n *null) Difference(other Interface) Interface { return n }

func (n *null) IsSubset(other Interface) bool { return true }

func (n *null) Equal(other Interface) bool { return other.IsEmpty() }

// IsNull reports that this is the null OrderedSet.
func (n *null) IsNull() bool { return true }
