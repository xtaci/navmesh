package main

import (
	"fmt"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	. "github.com/spate/vectormath"
	//. "github.com/xtaci/navmesh"
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

	triangles = [][3]int32{
		{1, 2, 3},
		{2, 3, 4},
		{3, 4, 5},
		{4, 5, 6},
		{5, 6, 7},
		{6, 7, 8},
		{6, 8, 10},
		{10, 8, 9},
		{10, 9, 11},
		{9, 11, 12},
		{9, 12, 14},
		{12, 13, 14},
		{13, 14, 16},
		{13, 15, 16},
		{14, 16, 17},
		{14, 17, 18},
		{16, 17, 19},
		{16, 20, 19},
		{17, 19, 21},
		{17, 21, 22},
	}
)

func main() {
	gl.StartDriver(appMain)
}

func appMain(driver gxui.Driver) {
	theme := flags.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "Polygon")
	window.SetScale(100)
	canvas := driver.CreateCanvas(math.Size{W: 1000, H: 1000})

	// mouse
	//var isStart bool
	//	var src_id, dest_id int // source & dest triangle id
	window.OnMouseDown(func(me gxui.MouseEvent) {
		getTriangleId(driver, me.Point, me.Point)
	})

	// draw mesh
	for k := 0; k < len(triangles); k++ {
		poly := []gxui.PolygonVertex{
			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][0]].X),
					int(vertices[triangles[k][0]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][1]].X),
					int(vertices[triangles[k][1]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][2]].X),
					int(vertices[triangles[k][2]].Y),
				},
				RoundedRadius: 0},
		}
		canvas.DrawPolygon(poly, gxui.CreatePen(0.01, gxui.Red), gxui.TransparentBrush)
	}

	canvas.Complete()
	image := theme.CreateImage()
	image.SetCanvas(canvas)
	window.AddChild(image)
	window.OnClose(driver.Terminate)
}

/*
func route(src, dest math.Point) {
	// Phase 0. Draw Triangle Id Graph

	// Phase 1. Use Dijkstra to find shortest path on Triangles
	d := Dijkstra{}
	d.CreateMatrixFromMesh(Mesh{vertices, triangles})
	path := d.Run(src_id)

	// Phase 2.  construct path indices
	// Check if this path include src & dest
	cur_id, ok := dest_id, false
	for cur_id, ok := path[cur_id]; ok; cur_id, ok = path[cur_id] {
		path_triangle = append(cur_id, path_triangle)
		if cur_id == src_id { // complete route
			break
		}
	}
	if cur_id != src_id { // incomplete route
		return
	}

	// Phase 3. use Navmesh to construct line
	start, end := &Point3{X: src.X, Y: src.Y}, &Point3{X: src.X, Y: src.Y}
	nm := NavMesh{}
	trilist := TriangleList{vertices, path_indices}
	r, _ := nm.Route(trilist, start, end)

	var poly []gxui.PolygonVertex
	poly = append(poly,
		gxui.PolygonVertex{
			Position: math.Point{
				int(start.X),
				int(start.Y),
			},
			RoundedRadius: 0})

	for k := range r.Line {
		poly = append(poly,
			gxui.PolygonVertex{
				Position: math.Point{
					int(r.Line[k].X),
					int(r.Line[k].Y),
				},
				RoundedRadius: 0})
	}
	poly = append(poly,
		gxui.PolygonVertex{
			Position: math.Point{
				int(end.X),
				int(end.Y),
			},
			RoundedRadius: 0})

	canvas.DrawLines(poly, gxui.CreatePen(0.01, gxui.Green))

}
*/

func getTriangleId(driver gxui.Driver, src, des math.Point) {
	canvas := driver.CreateCanvas(math.Size{W: 1000, H: 1000})
	canvas.Clear(gxui.Color{0, 0, 0, 1})

	// draw mesh with id
	for k := 0; k < len(triangles); k++ {
		poly := []gxui.PolygonVertex{
			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][0]].X),
					int(vertices[triangles[k][0]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][1]].X),
					int(vertices[triangles[k][1]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					int(vertices[triangles[k][2]].X),
					int(vertices[triangles[k][2]].Y),
				},
				RoundedRadius: 0},
		}
		color := gxui.Color{float32(k + 1), 0.0, 0.0, 1.0}
		canvas.DrawPolygon(poly, gxui.CreatePen(0.01, color), gxui.CreateBrush(color))
	}
	canvas.Complete()
	theme := flags.CreateTheme(driver)
	image := theme.CreateImage()
	image.SetCanvas(canvas)
	fmt.Println(image.Texture()) //.Image().At(src.X, src.Y))
}
