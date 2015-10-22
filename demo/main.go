package main

import (
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	. "github.com/spate/vectormath"
	. "github.com/xtaci/navmesh"
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

const SCALE_FACTOR = 100

func appMain(driver gxui.Driver) {
	theme := flags.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "Polygon")
	canvas := driver.CreateCanvas(math.Size{W: 800, H: 600})

	// mouse
	isStart := true
	var src_id, dest_id int32 // source & dest triangle id
	var src, dest Point3
	window.OnMouseDown(func(me gxui.MouseEvent) {
		pt := Point3{X: float32(me.Point.X) / SCALE_FACTOR, Y: float32(me.Point.Y) / SCALE_FACTOR}
		id := getTriangleId(pt)
		if id == -1 {
			return
		}
		if isStart {
			src_id = id
			src = pt
		} else {
			dest_id = id
			dest = pt
		}
		if !isStart {
			canvas := route(driver, src_id, dest_id, src, dest)
			image := theme.CreateImage()
			image.SetCanvas(canvas)
			window.AddChild(image)
		}
		isStart = !isStart
	})

	// draw mesh
	for k := 0; k < len(triangles); k++ {
		poly := []gxui.PolygonVertex{
			gxui.PolygonVertex{
				Position: math.Point{
					SCALE_FACTOR * int(vertices[triangles[k][0]].X),
					SCALE_FACTOR * int(vertices[triangles[k][0]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					SCALE_FACTOR * int(vertices[triangles[k][1]].X),
					SCALE_FACTOR * int(vertices[triangles[k][1]].Y),
				},
				RoundedRadius: 0},

			gxui.PolygonVertex{
				Position: math.Point{
					SCALE_FACTOR * int(vertices[triangles[k][2]].X),
					SCALE_FACTOR * int(vertices[triangles[k][2]].Y),
				},
				RoundedRadius: 0},
		}
		canvas.DrawPolygon(poly, gxui.CreatePen(1, gxui.Red), gxui.TransparentBrush)
	}

	canvas.Complete()
	image := theme.CreateImage()
	image.SetCanvas(canvas)
	window.AddChild(image)
	window.OnClose(driver.Terminate)
}

func route(driver gxui.Driver, src_id, dest_id int32, src, dest Point3) (canvas gxui.Canvas) {
	defer func() {
		canvas.Complete()
	}()
	canvas = driver.CreateCanvas(math.Size{W: 800, H: 600})
	// Phase 1. Use Dijkstra to find shortest path on Triangles
	d := Dijkstra{}
	d.CreateMatrixFromMesh(Mesh{vertices, triangles})
	path := d.Run(src_id)

	// Phase 2.  construct path indices
	// Check if this path include src & dest
	var path_triangle [][3]int32
	cur_id, ok := path[dest_id]
	for ; ok; cur_id, ok = path[cur_id] {
		path_triangle = append([][3]int32{triangles[cur_id]}, path_triangle...)
		if cur_id == src_id { // complete route
			break
		}
	}
	if cur_id != src_id { // incomplete route
		return canvas
	}

	// Phase 3. use Navmesh to construct line
	start, end := &Point3{X: src.X, Y: src.Y}, &Point3{X: dest.X, Y: dest.Y}
	nm := NavMesh{}
	trilist := TriangleList{vertices, path_triangle}
	r, _ := nm.Route(trilist, start, end)

	var poly []gxui.PolygonVertex
	poly = append(poly,
		gxui.PolygonVertex{
			Position: math.Point{
				int(SCALE_FACTOR * start.X),
				int(SCALE_FACTOR * start.Y),
			},
			RoundedRadius: 0})

	for k := range r.Line {
		poly = append(poly,
			gxui.PolygonVertex{
				Position: math.Point{
					int(SCALE_FACTOR * r.Line[k].X),
					int(SCALE_FACTOR * r.Line[k].Y),
				},
				RoundedRadius: 0})
	}
	poly = append(poly,
		gxui.PolygonVertex{
			Position: math.Point{
				int(SCALE_FACTOR * end.X),
				int(SCALE_FACTOR * end.Y),
			},
			RoundedRadius: 0})

	canvas.DrawLines(poly, gxui.CreatePen(1, gxui.Green))
	return
}

func sign(p1, p2, p3 Point3) float32 {
	return (p1.X-p3.X)*(p2.Y-p3.Y) - (p2.X-p3.X)*(p1.Y-p3.Y)
}

func inside(pt, v1, v2, v3 Point3) bool {
	b1 := sign(pt, v1, v2) <= 0
	b2 := sign(pt, v2, v3) <= 0
	b3 := sign(pt, v3, v1) <= 0
	return ((b1 == b2) && (b2 == b3))
}

func getTriangleId(pt Point3) (id int32) {
	for k := 0; k < len(triangles); k++ {
		if inside(pt,
			Point3{X: vertices[triangles[k][0]].X, Y: vertices[triangles[k][0]].Y},
			Point3{X: vertices[triangles[k][1]].X, Y: vertices[triangles[k][1]].Y},
			Point3{X: vertices[triangles[k][2]].X, Y: vertices[triangles[k][2]].Y}) {
			return int32(k)
		}
	}
	return -1
}

/*
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
}*/
