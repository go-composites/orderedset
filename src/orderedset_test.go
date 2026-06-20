package OrderedSet_test

import (
	Error "github.com/go-composites/error/src"
	OrderedSet "github.com/go-composites/orderedset/src"
	Result "github.com/go-composites/result/src"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// errResult builds a Result that reports HasError() == true, so Each
// short-circuits on it (HasError() is !error.IsNull()).
func errResult() Result.Interface {
	return Result.New(Result.WithError(Error.New("sentinel")))
}

// ints collects an OrderedSet's int items into a Go slice in iteration order.
// Because iteration is deterministic (insertion order), no sorting is needed —
// the assertions exercise the ordering directly.
func ints(s OrderedSet.Interface) []int {
	out := []int{}
	s.Each(func(item interface{}) Result.Interface {
		out = append(out, item.(int))
		return Result.New()
	})
	return out
}

var _ = ginkgo.Describe("OrderedSet", func() {
	ginkgo.Describe("New", func() {
		ginkgo.It("returns a non-nil, non-null, empty OrderedSet", func() {
			s := OrderedSet.New()
			gomega.Expect(s).NotTo(gomega.BeNil())
			gomega.Expect(s.IsNull()).To(gomega.BeFalse())
			gomega.Expect(s.Len()).To(gomega.Equal(0))
			gomega.Expect(s.IsEmpty()).To(gomega.BeTrue())
		})

		ginkgo.It("seeds variadic items, dedup preserving first-seen order", func() {
			s := OrderedSet.New(3, 1, 3, 2, 1, 2)
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{3, 1, 2}))
			gomega.Expect(s.IsEmpty()).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Add", func() {
		ginkgo.It("inserts items and returns the receiver for chaining", func() {
			s := OrderedSet.New()
			ret := s.Add(2).Add(1)
			gomega.Expect(ret).To(gomega.BeIdenticalTo(s))
			gomega.Expect(s.Len()).To(gomega.Equal(2))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{2, 1}))
		})

		ginkgo.It("is idempotent and preserves the original position", func() {
			s := OrderedSet.New(1, 2, 3).Add(1).Add(2)
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 2, 3}))
		})
	})

	ginkgo.Describe("Delete", func() {
		ginkgo.It("removes an item, preserving the remaining order", func() {
			s := OrderedSet.New(1, 2, 3, 4)
			ret := s.Delete(2)
			gomega.Expect(ret).To(gomega.BeIdenticalTo(s))
			gomega.Expect(s.Has(2)).To(gomega.BeFalse())
			gomega.Expect(s.Len()).To(gomega.Equal(3))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1, 3, 4}))
		})

		ginkgo.It("is a no-op for an absent item", func() {
			s := OrderedSet.New(1)
			s.Delete(99)
			gomega.Expect(s.Len()).To(gomega.Equal(1))
			gomega.Expect(ints(s)).To(gomega.Equal([]int{1}))
		})
	})

	ginkgo.Describe("Has", func() {
		ginkgo.It("is true for a present item", func() {
			s := OrderedSet.New(1)
			gomega.Expect(s.Has(1)).To(gomega.BeTrue())
		})

		ginkgo.It("is false for an absent item", func() {
			s := OrderedSet.New()
			gomega.Expect(s.Has(1)).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Each", func() {
		ginkgo.It("visits every item in insertion order, clean Result", func() {
			s := OrderedSet.New(3, 1, 2)
			out := []int{}
			res := s.Each(func(item interface{}) Result.Interface {
				out = append(out, item.(int))
				return Result.New()
			})
			gomega.Expect(out).To(gomega.Equal([]int{3, 1, 2}))
			gomega.Expect(res).NotTo(gomega.BeNil())
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
		})

		ginkgo.It("short-circuits on the first error Result", func() {
			s := OrderedSet.New(1, 2, 3)
			count := 0
			res := s.Each(func(item interface{}) Result.Interface {
				count++
				return errResult()
			})
			gomega.Expect(count).To(gomega.Equal(1))
			gomega.Expect(res.HasError()).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("ToArray", func() {
		ginkgo.It("materialises the items into an Array, in order", func() {
			s := OrderedSet.New(3, 1, 2)
			arr := s.ToArray()
			gomega.Expect(arr).NotTo(gomega.BeNil())
			gomega.Expect(arr.Len()).To(gomega.Equal(3))

			out := []int{}
			arr.Each(func(_ int, item interface{}) Result.Interface {
				out = append(out, item.(int))
				return Result.New()
			})
			gomega.Expect(out).To(gomega.Equal([]int{3, 1, 2}))
		})
	})

	ginkgo.Describe("Union", func() {
		ginkgo.It("keeps receiver order, then new items from the other", func() {
			a := OrderedSet.New(1, 2, 3)
			b := OrderedSet.New(3, 4, 5)
			gomega.Expect(ints(a.Union(b))).To(
				gomega.Equal([]int{1, 2, 3, 4, 5}))
			gomega.Expect(ints(b.Union(a))).To(
				gomega.Equal([]int{3, 4, 5, 1, 2}))
		})
	})

	ginkgo.Describe("Intersection", func() {
		ginkgo.It("keeps the common items in the receiver's order", func() {
			a := OrderedSet.New(1, 2, 3, 4)
			b := OrderedSet.New(4, 2)
			gomega.Expect(ints(a.Intersection(b))).To(
				gomega.Equal([]int{2, 4}))
			gomega.Expect(ints(b.Intersection(a))).To(
				gomega.Equal([]int{4, 2}))
		})
	})

	ginkgo.Describe("Difference", func() {
		ginkgo.It("keeps receiver items not in other, in order", func() {
			a := OrderedSet.New(1, 2, 3)
			b := OrderedSet.New(3, 4, 5)
			gomega.Expect(ints(a.Difference(b))).To(
				gomega.Equal([]int{1, 2}))
			gomega.Expect(ints(b.Difference(a))).To(
				gomega.Equal([]int{4, 5}))
		})
	})

	ginkgo.Describe("IsSubset", func() {
		ginkgo.It("is true when every item is in the other set", func() {
			gomega.Expect(OrderedSet.New(1, 2).IsSubset(
				OrderedSet.New(1, 2, 3))).To(gomega.BeTrue())
		})

		ginkgo.It("is false when an item is missing from the other set", func() {
			gomega.Expect(OrderedSet.New(1, 9).IsSubset(
				OrderedSet.New(1, 2, 3))).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Equal", func() {
		ginkgo.It("is order-INsensitive: same members in any order", func() {
			gomega.Expect(OrderedSet.New(1, 2, 3).Equal(
				OrderedSet.New(3, 2, 1))).To(gomega.BeTrue())
		})

		ginkgo.It("is false for sets of different sizes", func() {
			gomega.Expect(OrderedSet.New(1, 2).Equal(
				OrderedSet.New(1, 2, 3))).To(gomega.BeFalse())
		})

		ginkgo.It("is false for same-size sets with different items", func() {
			gomega.Expect(OrderedSet.New(1, 2, 3).Equal(
				OrderedSet.New(1, 2, 9))).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("Null", func() {
		ginkgo.It("is a Null-Object: IsNull true and inert", func() {
			n := OrderedSet.Null()
			gomega.Expect(n).NotTo(gomega.BeNil())
			gomega.Expect(n.IsNull()).To(gomega.BeTrue())
			gomega.Expect(n.Len()).To(gomega.Equal(0))
			gomega.Expect(n.IsEmpty()).To(gomega.BeTrue())

			// Mutators are no-ops that return the receiver.
			gomega.Expect(n.Add(1)).To(gomega.BeIdenticalTo(n))
			gomega.Expect(n.Delete(1)).To(gomega.BeIdenticalTo(n))
			gomega.Expect(n.Len()).To(gomega.Equal(0))

			// Membership always misses.
			gomega.Expect(n.Has(1)).To(gomega.BeFalse())

			// ToArray is empty.
			gomega.Expect(n.ToArray().Len()).To(gomega.Equal(0))

			// Set algebra returns the (inert) null set.
			gomega.Expect(n.Union(OrderedSet.New(1)).IsNull()).To(
				gomega.BeTrue())
			gomega.Expect(n.Intersection(OrderedSet.New(1)).IsNull()).To(
				gomega.BeTrue())
			gomega.Expect(n.Difference(OrderedSet.New(1)).IsNull()).To(
				gomega.BeTrue())

			// IsSubset of anything is true; Equal holds only for empty sets.
			gomega.Expect(n.IsSubset(OrderedSet.New(1))).To(gomega.BeTrue())
			gomega.Expect(n.Equal(OrderedSet.New())).To(gomega.BeTrue())
			gomega.Expect(n.Equal(OrderedSet.New(1))).To(gomega.BeFalse())

			// Each returns a clean Result without invoking fn.
			called := false
			res := n.Each(func(item interface{}) Result.Interface {
				called = true
				return errResult()
			})
			gomega.Expect(called).To(gomega.BeFalse())
			gomega.Expect(res.HasError()).To(gomega.BeFalse())
		})
	})
})
