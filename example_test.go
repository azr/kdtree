package kdtree

import (
	"fmt"
	"math/rand"
)

// Generate random points in the unit square, and prints all points
// within a radius of 0.25 and the 0.5 from the origin.
func ExampleT_InRange() {
	// Make a K-D tree of random points.
	const N = 1000
	nodes := make([]*T, N)
	for i := range nodes {
		nodes[i] = new(T)
		for j := range nodes[i].Point {
			nodes[i].Point[j] = rand.Float64()
		}
	}
	tree := New(nodes)

	nodes = tree.InRange(Point{0, 0}, 0.25, make([]*T, 0, N))
	fmt.Println(nodes)

	// Reuse the nodes slice from the previous call.
	nodes = tree.InRange(Point{0, 0}, 0.5, nodes[:0])
	fmt.Println(nodes)
}
