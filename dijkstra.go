package navmesh

import (
	"container/heap"
	. "github.com/spate/vectormath"
	"math"
)

import ()

// Triangle Heap
type WeightedTriangle struct {
	id     int // triangle id
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

type Mesh struct {
	Vertices []Point3 // triangle list, index start from 0
	Indices  []int
}

// Dijkstra
type Dijkstra struct {
	Matrix map[int][]WeightedTriangle // all edge for nodes
}

// create neighbour matrix
func (d *Dijkstra) CreateMatrixFromMesh(mesh Mesh) {
	d.Matrix = make(map[int][]WeightedTriangle)
	for i := 0; i < len(mesh.Indices); i += 3 {
		for j := 0; j < len(mesh.Indices); j += 3 {
			if i == j {
				continue
			}

			if len(intersect(mesh.Indices[i:i+3], mesh.Indices[j:j+3])) == 2 {
				x1 := (mesh.Vertices[mesh.Indices[i]].X + mesh.Vertices[mesh.Indices[i+1]].X + mesh.Vertices[mesh.Indices[i+2]].X) / 3.0
				y1 := (mesh.Vertices[mesh.Indices[i]].Y + mesh.Vertices[mesh.Indices[i+1]].Y + mesh.Vertices[mesh.Indices[i+2]].Y) / 3.0
				x2 := (mesh.Vertices[mesh.Indices[j]].X + mesh.Vertices[mesh.Indices[j+1]].X + mesh.Vertices[mesh.Indices[j+2]].X) / 3.0
				y2 := (mesh.Vertices[mesh.Indices[j]].Y + mesh.Vertices[mesh.Indices[j+1]].Y + mesh.Vertices[mesh.Indices[j+2]].Y) / 3.0
				weight := float32(math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))))
				d.Matrix[i] = append(d.Matrix[i], WeightedTriangle{j, weight})
			}
		}
	}
}

func intersect(a []int, b []int) []int {
	var inter []int
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				inter = append(inter, a[i])
			}
		}
	}
	return inter
}

func (d *Dijkstra) Run(src_id int) map[int]int {
	// triangle heap
	h := &TriangleHeap{}
	heap.Init(h)

	// min distance records
	dist := make(map[int]float32)

	// previous map
	prev := make(map[int]int)

	// visit map
	visited := make(map[int]bool)

	// set initial distance to each node as MaxFloat32
	for k := range d.Matrix {
		dist[k] = math.MaxFloat32
	}

	// source vertex, the first vertex in Heap
	heap.Push(h, WeightedTriangle{src_id, 0})
	dist[src_id] = 0.0

	for h.Len() > 0 { // for every un-visited vertex, try relaxing the path
		// pop the min element
		cur := h.Pop().(WeightedTriangle)
		if visited[cur.id] {
			continue
		}
		// current known shortest distance to u
		dist_u := dist[cur.id]
		// mark the vertex as visited.
		visited[cur.id] = true

		// for each neighbor v of u:
		for _, v := range d.Matrix[cur.id] {
			alt := dist_u + v.weight
			if alt < dist[v.id] && !visited[v.id] {
				dist[v.id] = alt
				prev[v.id] = cur.id
				heap.Push(h, WeightedTriangle{v.id, alt})
			}
		}
	}
	return prev
}
