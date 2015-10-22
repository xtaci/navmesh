package main

import (
	"encoding/json"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"
	. "github.com/spate/vectormath"
	. "github.com/xtaci/navmesh"
	"log"
	"os"
)

type List struct {
	Vertices  []Point3
	Triangles [][3]int32
}

var list List
var vertices []Point3
var triangles [][3]int32
var dijkstra Dijkstra

func main() {
	f, err := os.Open("mesh.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		log.Fatal(err)
	}
	vertices = list.Vertices
	triangles = list.Triangles
	dijkstra.CreateMatrixFromMesh(Mesh{vertices, triangles})
	gl.StartDriver(appMain)
}

var SCALE_FACTOR = float32(0.15)

func appMain(driver gxui.Driver) {
	theme := flags.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "navmesh")
	canvas := driver.CreateCanvas(math.Size{W: 800, H: 600})

	// mouse
	isStart := true
	var src_id, dest_id int32 // source & dest triangle id
	var src, dest Point3
	window.OnMouseDown(func(me gxui.MouseEvent) {
		pt := Point3{X: float32(me.Point.X) / SCALE_FACTOR, Y: float32(me.Point.Y) / SCALE_FACTOR}
		id := getTriangleId(pt)
		if isStart {
			src_id = id
			src = pt
		} else {
			dest_id = id
			dest = pt
		}
		if !isStart {
			if id != -1 {
				canvas := route(driver, src_id, dest_id, src, dest)
				image := theme.CreateImage()
				image.SetCanvas(canvas)
				window.AddChild(image)
			}
		}
		isStart = !isStart
	})

	// draw mesh
	for k := 0; k < len(triangles); k++ {
		poly := []gxui.PolygonVertex{
			gxui.PolygonVertex{
				Position: math.Point{
					int(SCALE_FACTOR * vertices[triangles[k][0]].X),
					int(SCALE_FACTOR * vertices[triangles[k][0]].Y),
				}},

			gxui.PolygonVertex{
				Position: math.Point{
					int(SCALE_FACTOR * vertices[triangles[k][1]].X),
					int(SCALE_FACTOR * vertices[triangles[k][1]].Y),
				}},

			gxui.PolygonVertex{
				Position: math.Point{
					int(SCALE_FACTOR * vertices[triangles[k][2]].X),
					int(SCALE_FACTOR * vertices[triangles[k][2]].Y),
				}},
		}
		canvas.DrawPolygon(poly, gxui.CreatePen(3, gxui.Gray80), gxui.CreateBrush(gxui.Gray40))
		//canvas.DrawPolygon(poly, gxui.CreatePen(2, gxui.Red), gxui.CreateBrush(gxui.Yellow))
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
	path := dijkstra.Run(src_id)

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
	if cur_id != src_id && src_id != dest_id { // incomplete route
		return canvas
	}
	log.Println(path_triangle)

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
			}})

	for k := range r.Line {
		poly = append(poly,
			gxui.PolygonVertex{
				Position: math.Point{
					int(SCALE_FACTOR * r.Line[k].X),
					int(SCALE_FACTOR * r.Line[k].Y),
				}})
	}
	poly = append(poly,
		gxui.PolygonVertex{
			Position: math.Point{
				int(SCALE_FACTOR * end.X),
				int(SCALE_FACTOR * end.Y),
			}})

	canvas.DrawLines(poly, gxui.CreatePen(2, gxui.Green))
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
