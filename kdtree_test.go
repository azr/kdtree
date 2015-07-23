package kdtree

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// A pointSlice is a slice of points that implements the quick.Generator
// interface, generating a random set of points on the unit square.
type pointSlice []Point

func (pointSlice) Generate(r *rand.Rand, size int) reflect.Value {
	ps := make([]Point, size)
	for i := range ps {
		for j := range ps[i] {
			ps[i][j] = r.Float64()
		}
	}
	return reflect.ValueOf(ps)
}

// Generate implements the Generator interface for Points
func (p Point) Generate(r *rand.Rand, _ int) reflect.Value {
	for i := range p {
		p[i] = r.Float64()
	}
	return reflect.ValueOf(p)
}

// TestInsert tests the insert function, ensuring that random points
// inserted into an empty tree maintain the K-D tree invariant.
func TestInsert(t *testing.T) {
	if err := quick.Check(func(pts pointSlice) bool {
		var tree *T
		for _, p := range pts {
			tree = tree.Insert(&T{Point: p})
		}
		_, ok := tree.invariantHolds()
		return ok
	}, nil); err != nil {
		t.Error(err)
	}
}

// When splitting on the median, values after the median index may tie the median.
// Make sure that we correctly account for this when splitting and reusing slices.
// The test ensures that we don't panic.
// See issue 18.
func TestMedianTies(t *testing.T) {
	 New([]*T{
		&T{Point: Point{0, 1}},
		&T{Point: Point{0, 1}},
		&T{Point: Point{0, 0}},
		&T{Point: Point{1, 0}},
	})

	// This is the data that originally produced issue 18.
	 New([]*T{
		&T{Point: Point{6, 9}},
		&T{Point: Point{6, 4}},
		&T{Point: Point{4, 6}},
		&T{Point: Point{0, 1}},
		&T{Point: Point{0, 3}},
		&T{Point: Point{5, 8}},
		&T{Point: Point{2, 3}},
		&T{Point: Point{3, 4}},
		&T{Point: Point{2, 2}},
		&T{Point: Point{6, 2}},
		&T{Point: Point{2, 3}},
		&T{Point: Point{5, 8}},
		&T{Point: Point{2, 2}},
		&T{Point: Point{7, 2}},
		&T{Point: Point{8, 6}},
		&T{Point: Point{5, 0}},
		&T{Point: Point{1, 6}},
		&T{Point: Point{9, 0}},
	})
}

// TestMake tests the Make function, ensuring that a tree built
// using random points respects the K-D tree invariant.
func TestMake(t *testing.T) {
	if err := quick.Check(func(pts pointSlice) bool {
		nodes := make([]*T, len(pts))
		for i, pt := range pts {
			nodes[i] = &T{Point: pt}
		}
		tree := New(nodes)
		_, ok := tree.invariantHolds()
		return ok
	}, nil); err != nil {
		t.Error(err)
	}
}

// TestInRange tests the InRange function, ensuring that all points
// in the range are reported, and all points reported are indeed in
// the range.
func TestInRange(t *testing.T) {
	if err := quick.Check(func(pts pointSlice, pt Point, r float64) bool {
		r = math.Abs(r)
		nodes := make([]*T, len(pts))
		for i, pt := range pts {
			nodes[i] = &T{Point: pt}
		}

		tree := New(nodes)
		in := make(map[*T]bool, len(nodes))
		for _, n := range tree.InRange(pt, r, nil) {
			in[n] = true
		}

		num := 0
		for _, n := range nodes {
			if pt.sqDist(&n.Point) <= r*r {
				num++
				if !in[n] {
					return false
				}
			}
		}
		return num == len(in)
	}, nil); err != nil {
		t.Error(err)
	}
}

// InvariantHolds returns the points in this subtree, and a bool
// that is true if the K-D tree invariant holds.  The K-D tree invariant
// states that all points in the left subtree have values less than that
// of the current node on the splitting dimension, and the points
// in the right subtree have values greater than or equal to that of
// the current node.
func (t *T) invariantHolds() ([]Point, bool) {
	if t == nil {
		return []Point{}, true
	}

	left, leftOk := t.left.invariantHolds()
	right, rightOk := t.right.invariantHolds()

	ok := leftOk && rightOk

	if ok {
		for _, l := range left {
			if l[t.split] >= t.Point[t.split] {
				ok = false
				break
			}
		}
	}
	if ok {
		for _, r := range right {
			if r[t.split] < t.Point[t.split] {
				ok = false
				break
			}
		}
	}
	return append(append(left, t.Point), right...), ok
}

func TestPreSort(t *testing.T) {
	if err := quick.Check(func(pts pointSlice) bool {
		nodes := make([]*T, len(pts))
		for i, pt := range pts {
			nodes[i] = &T{Point: pt}
		}

		p := preSort(nodes)
		for i := range p.cur {
			if !isSortedOnDim(i, p.cur[i]) || len(p.cur[i]) != len(nodes) {
				return false
			}
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestPreSort_SplitMed(t *testing.T) {
	if err := quick.Check(func(pts pointSlice, dim int) bool {
		if len(pts) == 0 {
			return true
		}
		if dim < 0 {
			dim = -dim
		}
		dim %= K

		nodes := make([]*T, len(pts))
		for i, pt := range pts {
			nodes[i] = &T{Point: pt}
		}

		sorted := preSort(nodes)
		med, left, right := sorted.splitMed(dim)

		for i, p := range [2]*preSorted{&left, &right} {
			for d, ns := range p.cur {
				if len(ns) != p.Len() {
					return false
				}
				if !isSortedOnDim(d, ns) {
					return false
				}
				for _, n := range ns {
					if i == 0 && n.Point[dim] >= med.Point[dim] {
						return false
					} else if i == 1 && n.Point[dim] < med.Point[dim] {
						return false
					}
				}
			}
		}

		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

// IsSortedOnDim returns true if the given slice is in sorted order
// on the given dimension.
func isSortedOnDim(dim int, nodes []*T) bool {
	if len(nodes) == 0 {
		return true
	}
	prev := nodes[0].Point[dim]
	for _, n := range nodes {
		if n.Point[dim] < prev {
			return false
		}
		prev = n.Point[dim]
	}
	return true
}
