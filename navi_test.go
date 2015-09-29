package navmesh

import (
	"fmt"
	. "misc/vectormath"
	"testing"
)

func TestVectormath(t *testing.T) {
	v1 := Vector3{X: 1, Y: 1, Z: 0}
	v2 := Vector3{X: 1, Y: 0.5, Z: 0}

	res := Vector3{}
	V3Cross(&res, &v1, &v2)
	fmt.Println(res)
}

var (
	vertices = []Point3{{},
		{X: 0, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 2, Y: 0},
		{X: 2, Y: 1},
		{X: 3, Y: 0},
		{X: 3, Y: 1},
		{X: 3, Y: 2},
		{X: 2, Y: 2},
		{X: 2, Y: 3},
		{X: 3, Y: 3},
	}
	indices = []int{1, 2, 3, 2, 3, 4, 3, 4, 5, 4, 5, 6, 5, 6, 7, 6, 7, 8, 6, 8, 10, 10, 8, 9} //,10, 9, 11, 9, 11, 12}
)

func TestNavi(t *testing.T) {
	nm := NavMesh{}
	trilist := TriangleList{vertices, indices}
	r, _ := nm.Route(trilist, &Point3{X: 0.2, Y: 0.2}, &Point3{X: 2.5, Y: 2.999})
	fmt.Println("route:", r)

	r, _ = nm.Route(trilist, &Point3{X: 0.2, Y: 0.2}, &Point3{X: 3, Y: 1.01})
	fmt.Println("route:", r)
}

func BenchmarkNavi(b *testing.B) {
	nm := NavMesh{}
	trilist := TriangleList{vertices, indices}
	for i := 0; i < b.N; i++ {
		nm.Route(trilist, &Point3{X: 0.2, Y: 0.2}, &Point3{X: 2.5, Y: 2.999})
	}
}
