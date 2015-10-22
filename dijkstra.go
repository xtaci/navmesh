package navmesh

import (
	"container/heap"
	. "github.com/spate/vectormath"
	"math"
)

import ()

// Triangle Heap
type WeightedTriangle struct {
	id     int32 // triangle id
	weight float32
}

type TriangleHeap struct {
	triangles []WeightedTriangle
}

func (th *TriangleHeap) Len() int {
	return len(th.triangles)
}

func (th *TriangleHeap) Less(i, j int) bool {
	return th.triangles[i].weight < th.triangles[j].weight
}

func (th *TriangleHeap) Swap(i, j int) {
	th.triangles[i], th.triangles[j] = th.triangles[j], th.triangles[i]
}

func (th *TriangleHeap) Push(x interface{}) {
	th.triangles = append(th.triangles, x.(WeightedTriangle))
}

func (th *TriangleHeap) Pop() interface{} {
	n := len(th.triangles)
	x := th.triangles[n-1]
	th.triangles = th.triangles[:n-1]
	return x
}

func (th *TriangleHeap) DecreaseKey(id int32, weight float32) {
	for k := range th.triangles {
		if th.triangles[k].id == id {
			th.triangles[k].weight = weight
			heap.Fix(th, k)
			println("fixed", id)
			return
		}
	}
}

type Mesh struct {
	Vertices  []Point3   // vertices
	Triangles [][3]int32 // triangles
}

// Dijkstra
type Dijkstra struct {
	Matrix map[int32][]WeightedTriangle // all edge for nodes
}

// create neighbour matrix
func (d *Dijkstra) CreateMatrixFromMesh(mesh Mesh) {
	d.Matrix = make(map[int32][]WeightedTriangle)
	for i := 0; i < len(mesh.Triangles); i++ {
		for j := 0; j < len(mesh.Triangles); j++ {
			if i == j {
				continue
			}

			if len(intersect(mesh.Triangles[i], mesh.Triangles[j])) == 2 {
				x1 := (mesh.Vertices[mesh.Triangles[i][0]].X + mesh.Vertices[mesh.Triangles[i][1]].X + mesh.Vertices[mesh.Triangles[i][2]].X) / 3.0
				y1 := (mesh.Vertices[mesh.Triangles[i][0]].Y + mesh.Vertices[mesh.Triangles[i][1]].Y + mesh.Vertices[mesh.Triangles[i][2]].Y) / 3.0
				x2 := (mesh.Vertices[mesh.Triangles[j][0]].X + mesh.Vertices[mesh.Triangles[j][1]].X + mesh.Vertices[mesh.Triangles[j][2]].X) / 3.0
				y2 := (mesh.Vertices[mesh.Triangles[j][0]].Y + mesh.Vertices[mesh.Triangles[j][1]].Y + mesh.Vertices[mesh.Triangles[j][2]].Y) / 3.0
				weight := float32(math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))))
				d.Matrix[int32(i)] = append(d.Matrix[int32(i)], WeightedTriangle{int32(j), weight})
			}
		}
	}
}

func intersect(a [3]int32, b [3]int32) []int32 {
	var inter []int32
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				inter = append(inter, a[i])
			}
		}
	}
	return inter
}

func (d *Dijkstra) Run(src_id int32) map[int32]int32 {
	// triangle heap
	h := &TriangleHeap{}
	heap.Init(h)

	// min distance records
	dist := make(map[int32]float32)

	// previous map
	prev := make(map[int32]int32)

	// visit map
	visited := make(map[int32]bool)

	// set initial distance to a very large value
	for k := range d.Matrix {
		dist[k] = 99999
	}

	// source vertex, the first vertex in Heap
	heap.Push(h, WeightedTriangle{src_id, 0})
	dist[src_id] = 0.0

	for h.Len() > 0 { // for every un-visited vertex, try relaxing the path
		// pop the min element
		u := h.Pop().(WeightedTriangle)
		// current known shortest distance to u
		dist_u := dist[u.id]
		// mark the vertex as visited.
		visited[u.id] = true

		// for each neighbor v of u:
		for _, v := range d.Matrix[u.id] {
			alt := dist_u + v.weight // from src->u->v
			if !visited[v.id] {
				heap.Push(h, WeightedTriangle{v.id, alt})
			}
			if alt < dist[v.id] {
				dist[v.id] = alt
				prev[v.id] = u.id
				if !visited[v.id] {
					h.DecreaseKey(v.id, alt)
				}
			}
		}
	}
	return prev
}
