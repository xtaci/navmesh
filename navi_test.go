package navmesh

import (
	. "github.com/spate/vectormath"
	"github.com/veandco/go-sdl2/sdl"
	"testing"
	"time"
)

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
		{X: 4, Y: 3},
		{X: 4, Y: 2}, //14
		{X: 5, Y: 3},
		{X: 5, Y: 2},
		{X: 5, Y: 1},
		{X: 4, Y: 1},
		{X: 6, Y: 1},
		{X: 6, Y: 2},
		{X: 6, Y: 0},
		{X: 5, Y: 0},
	}
	indices = []int{1, 2, 3, 2, 3, 4, 3, 4, 5, 4, 5, 6, 5, 6, 7, 6, 7, 8, 6, 8, 10, 10, 8, 9, 10, 9, 11, 9, 11, 12, 9, 12, 14, 12, 13, 14,
		13, 14, 16,
		13, 15, 16,
		14, 16, 17,
		14, 17, 18,
		16, 17, 19,
		16, 20, 19,
		17, 19, 21,
		17, 21, 22,
	}
)

func TestNavmesh(t *testing.T) {
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateSoftwareRenderer(surface)
	if err != nil {
		panic(err)
	}

	renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff)
	C := float32(100)

	for k := 0; k < len(indices); k += 3 {
		renderer.DrawLine(int(C*vertices[indices[k]].X), int(C*vertices[indices[k]].Y), int(C*vertices[indices[k+1]].X), int(C*vertices[indices[k+1]].Y))
		renderer.DrawLine(int(C*vertices[indices[k+1]].X), int(C*vertices[indices[k+1]].Y), int(C*vertices[indices[k+2]].X), int(C*vertices[indices[k+2]].Y))
		renderer.DrawLine(int(C*vertices[indices[k+2]].X), int(C*vertices[indices[k+2]].Y), int(C*vertices[indices[k]].X), int(C*vertices[indices[k]].Y))
	}
	d := Dijkstra{}
	d.CreateMatrixFromMesh(Mesh{vertices, indices})
	path := d.Run(0)
	cur, ok := 57, false
	renderer.SetDrawColor(0x00, 0xff, 0x00, 0xff)
	var path_indices []int
	for {
		renderer.DrawLine(int(C*vertices[indices[cur]].X), int(C*vertices[indices[cur]].Y), int(C*vertices[indices[cur+1]].X), int(C*vertices[indices[cur+1]].Y))
		renderer.DrawLine(int(C*vertices[indices[cur+1]].X), int(C*vertices[indices[cur+1]].Y), int(C*vertices[indices[cur+2]].X), int(C*vertices[indices[cur+2]].Y))
		renderer.DrawLine(int(C*vertices[indices[cur+2]].X), int(C*vertices[indices[cur+2]].Y), int(C*vertices[indices[cur]].X), int(C*vertices[indices[cur]].Y))
		path_indices = append(indices[cur:cur+3], path_indices...)
		if cur, ok = path[cur]; !ok {
			break
		}
	}

	// construct path indices
	start, end := &Point3{X: 0.2, Y: 0.2}, &Point3{X: 5.1, Y: 0.2}
	nm := NavMesh{}
	trilist := TriangleList{vertices, path_indices}
	r, _ := nm.Route(trilist, start, end)
	sdl_line := []sdl.Point{{X: int32(C * start.X), Y: int32(C * start.Y)}}
	for k := range r.Line {
		sdl_line = append(sdl_line, sdl.Point{X: int32(C * r.Line[k].X), Y: int32(C * r.Line[k].Y)})
	}
	sdl_line = append(sdl_line, sdl.Point{X: int32(C * end.X), Y: int32(C * end.Y)})
	renderer.SetDrawColor(0xff, 0x00, 0x00, 0xff)
	renderer.DrawLines(sdl_line)

	window.UpdateSurface()
	<-time.After(time.Minute)
	sdl.Quit()
}
