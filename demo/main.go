package main

import (
	"encoding/json"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	. "github.com/spate/vectormath"
	. "github.com/xtaci/navmesh"
)

const (
	windowWidth  = 800
	windowHeight = 600
	SCALE_FACTOR = float32(0.15)
)

type List struct {
	Vertices          []Point3
	Triangles         [][3]int32
	ObstacleTriangles []int32 `json:"ObstacleTriangles"`
}

var (
	walkableFill = color.NRGBA{R: 72, G: 160, B: 255, A: 210}
	obstacleFill = color.NRGBA{R: 255, G: 110, B: 92, A: 255}
	edgeColor    = color.NRGBA{R: 90, G: 96, B: 120, A: 255}
)

func main() {
	meshFile, err := os.Open("mesh.json")
	if err != nil {
		log.Fatal(err)
	}
	defer meshFile.Close()

	var list List
	if err := json.NewDecoder(meshFile).Decode(&list); err != nil {
		log.Fatal(err)
	}

	obstacleSet := make(map[int32]struct{}, len(list.ObstacleTriangles))
	for _, id := range list.ObstacleTriangles {
		if id >= 0 {
			obstacleSet[id] = struct{}{}
		}
	}

	dijkstra := &Dijkstra{}
	dijkstra.CreateMatrixFromMesh(Mesh{Vertices: list.Vertices, Triangles: list.Triangles, Obstacles: obstacleSet})

	application := app.New()
	window := application.NewWindow("navmesh")
	window.Resize(fyne.NewSize(windowWidth, windowHeight))

	navWidget := newNavMeshWidget(list.Vertices, list.Triangles, obstacleSet, dijkstra)
	window.SetContent(navWidget)
	window.ShowAndRun()
}

type navMeshWidget struct {
	widget.BaseWidget
	container      *fyne.Container
	background     *canvas.Rectangle
	vertices       []Point3
	triangles      [][3]int32
	obstacles      map[int32]struct{}
	dijkstra       *Dijkstra
	selectingStart bool
	startPoint     Point3
	destPoint      Point3
	startTriangle  int32
	destTriangle   int32
	markerObjects  []fyne.CanvasObject
	pathObjects    []fyne.CanvasObject
}

func newNavMeshWidget(vertices []Point3, triangles [][3]int32, obstacles map[int32]struct{}, dijkstra *Dijkstra) *navMeshWidget {
	bg := canvas.NewRectangle(color.NRGBA{R: 18, G: 18, B: 26, A: 255})
	content := container.NewWithoutLayout(bg)
	w := &navMeshWidget{
		container:      content,
		background:     bg,
		vertices:       vertices,
		triangles:      triangles,
		obstacles:      obstacles,
		dijkstra:       dijkstra,
		selectingStart: true,
		startTriangle:  -1,
		destTriangle:   -1,
	}
	w.ExtendBaseWidget(w)
	w.drawMesh()
	return w
}

func (n *navMeshWidget) CreateRenderer() fyne.WidgetRenderer {
	return &navMeshRenderer{
		widget:  n,
		objects: []fyne.CanvasObject{n.container},
	}
}

type navMeshRenderer struct {
	widget  *navMeshWidget
	objects []fyne.CanvasObject
}

func (r *navMeshRenderer) Destroy() {}

func (r *navMeshRenderer) Layout(size fyne.Size) {
	r.widget.container.Resize(size)
	r.widget.background.Resize(size)
}

func (r *navMeshRenderer) MinSize() fyne.Size {
	return fyne.NewSize(windowWidth, windowHeight)
}

func (r *navMeshRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *navMeshRenderer) Refresh() {
	canvas.Refresh(r.widget.container)
}

func (r *navMeshRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (n *navMeshWidget) drawMesh() {
	for idx, tri := range n.triangles {
		p0 := fyne.NewPos(SCALE_FACTOR*n.vertices[tri[0]].X, SCALE_FACTOR*n.vertices[tri[0]].Y)
		p1 := fyne.NewPos(SCALE_FACTOR*n.vertices[tri[1]].X, SCALE_FACTOR*n.vertices[tri[1]].Y)
		p2 := fyne.NewPos(SCALE_FACTOR*n.vertices[tri[2]].X, SCALE_FACTOR*n.vertices[tri[2]].Y)
		fill := walkableFill
		if _, blocked := n.obstacles[int32(idx)]; blocked {
			fill = obstacleFill
		}
		img := newTriangleImage([3]fyne.Position{p0, p1, p2}, fill)
		n.container.Add(img)
		drawTriangleOutline(n.container, [3]fyne.Position{p0, p1, p2})
	}
}

func newTriangleImage(pts [3]fyne.Position, fill color.Color) *canvas.Image {
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := pts[0].X, pts[0].Y
	for i := 1; i < 3; i++ {
		if pts[i].X < minX {
			minX = pts[i].X
		}
		if pts[i].Y < minY {
			minY = pts[i].Y
		}
		if pts[i].X > maxX {
			maxX = pts[i].X
		}
		if pts[i].Y > maxY {
			maxY = pts[i].Y
		}
	}
	width := int(math.Ceil(float64(maxX-minX))) + 1
	height := int(math.Ceil(float64(maxY-minY))) + 1
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	local := [3]fyne.Position{
		fyne.NewPos(pts[0].X-minX, pts[0].Y-minY),
		fyne.NewPos(pts[1].X-minX, pts[1].Y-minY),
		fyne.NewPos(pts[2].X-minX, pts[2].Y-minY),
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	fillNRGBA := color.NRGBAModel.Convert(fill).(color.NRGBA)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if pointInTriangle(float32(x)+0.5, float32(y)+0.5, local) {
				img.Set(x, y, fillNRGBA)
			} else {
				img.Set(x, y, color.NRGBA{0, 0, 0, 0})
			}
		}
	}
	canvasImg := canvas.NewImageFromImage(img)
	canvasImg.Move(fyne.NewPos(minX, minY))
	size := fyne.NewSize(float32(width), float32(height))
	canvasImg.SetMinSize(size)
	canvasImg.Resize(size)
	canvasImg.FillMode = canvas.ImageFillOriginal
	return canvasImg
}

func drawTriangleOutline(c *fyne.Container, pts [3]fyne.Position) {
	for i := 0; i < 3; i++ {
		next := (i + 1) % 3
		line := canvas.NewLine(edgeColor)
		line.StrokeWidth = 1
		line.Position1 = pts[i]
		line.Position2 = pts[next]
		c.Add(line)
	}
}

func pointInTriangle(px, py float32, pts [3]fyne.Position) bool {
	b1 := sameSide(px, py, pts[0], pts[1], pts[2])
	b2 := sameSide(px, py, pts[1], pts[2], pts[0])
	b3 := sameSide(px, py, pts[2], pts[0], pts[1])
	return b1 && b2 && b3
}

func sameSide(px, py float32, a, b, c fyne.Position) bool {
	side1 := edgeSign(px, py, a, b)
	side2 := edgeSign(float32(c.X), float32(c.Y), a, b)
	if side2 == 0 {
		return true
	}
	if side2 > 0 {
		return side1 >= 0
	}
	return side1 <= 0
}

func edgeSign(px, py float32, a, b fyne.Position) float32 {
	ax := float32(a.X)
	ay := float32(a.Y)
	bx := float32(b.X)
	by := float32(b.Y)
	return (px-bx)*(ay-by) - (ax-bx)*(py-by)
}

func (n *navMeshWidget) Tapped(ev *fyne.PointEvent) {
	world := Point3{X: float32(ev.Position.X) / SCALE_FACTOR, Y: float32(ev.Position.Y) / SCALE_FACTOR}
	tri := n.triangleAt(world)
	if tri == -1 {
		return
	}
	if _, blocked := n.obstacles[tri]; blocked {
		return
	}

	if n.selectingStart {
		n.startPoint = world
		n.startTriangle = tri
		n.destTriangle = -1
		n.selectingStart = false
		n.clearObjects(&n.pathObjects)
		n.updateMarkers()
		return
	}

	n.destPoint = world
	n.destTriangle = tri
	n.selectingStart = true
	n.updateMarkers()
	n.drawPath()
}

func (n *navMeshWidget) triangleAt(pt Point3) int32 {
	for i, tri := range n.triangles {
		if insideTriangle(pt,
			Point3{X: n.vertices[tri[0]].X, Y: n.vertices[tri[0]].Y},
			Point3{X: n.vertices[tri[1]].X, Y: n.vertices[tri[1]].Y},
			Point3{X: n.vertices[tri[2]].X, Y: n.vertices[tri[2]].Y}) {
			return int32(i)
		}
	}
	return -1
}

func (n *navMeshWidget) updateMarkers() {
	n.clearObjects(&n.markerObjects)
	if n.startTriangle != -1 {
		n.markerObjects = append(n.markerObjects, n.addMarker(n.startPoint, color.NRGBA{R: 220, G: 200, B: 60, A: 255}))
	}
	if n.destTriangle != -1 {
		n.markerObjects = append(n.markerObjects, n.addMarker(n.destPoint, color.NRGBA{R: 90, G: 180, B: 255, A: 255}))
	}
	n.Refresh()
}

func (n *navMeshWidget) addMarker(pt Point3, col color.Color) fyne.CanvasObject {
	circle := canvas.NewCircle(col)
	size := float32(10)
	circle.Resize(fyne.NewSize(size, size))
	circle.Move(fyne.NewPos(SCALE_FACTOR*pt.X-size/2, SCALE_FACTOR*pt.Y-size/2))
	n.container.Add(circle)
	return circle
}

func (n *navMeshWidget) drawPath() {
	n.clearObjects(&n.pathObjects)
	if n.startTriangle == -1 || n.destTriangle == -1 {
		n.Refresh()
		return
	}

	points := n.computeRoutePoints(n.startTriangle, n.destTriangle, n.startPoint, n.destPoint)
	if len(points) < 2 {
		n.Refresh()
		return
	}

	pathColor := color.NRGBA{R: 60, G: 210, B: 130, A: 255}
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(pathColor)
		line.StrokeWidth = 3
		line.Position1 = fyne.NewPos(SCALE_FACTOR*points[i].X, SCALE_FACTOR*points[i].Y)
		line.Position2 = fyne.NewPos(SCALE_FACTOR*points[i+1].X, SCALE_FACTOR*points[i+1].Y)
		n.container.Add(line)
		n.pathObjects = append(n.pathObjects, line)
	}
	n.Refresh()
}

func (n *navMeshWidget) computeRoutePoints(srcID, destID int32, src, dest Point3) []Point3 {
	if srcID < 0 || destID < 0 || int(srcID) >= len(n.triangles) || int(destID) >= len(n.triangles) {
		return nil
	}
	if _, blocked := n.obstacles[srcID]; blocked {
		return nil
	}
	if _, blocked := n.obstacles[destID]; blocked {
		return nil
	}

	path := n.dijkstra.Run(srcID)
	if int(destID) >= len(path) {
		return nil
	}

	pathTriangles := [][3]int32{n.triangles[destID]}
	prev := destID
	for {
		cur := path[prev]
		if cur == -1 {
			return nil
		}
		pathTriangles = append([][3]int32{n.triangles[cur]}, pathTriangles...)
		if cur == srcID {
			break
		}
		prev = cur
		if prev < 0 || int(prev) >= len(path) {
			return nil
		}
	}

	startCopy := src
	endCopy := dest
	nm := NavMesh{}
	r, err := nm.Route(TriangleList{Vertices: n.vertices, Triangles: pathTriangles}, &startCopy, &endCopy)
	if err != nil {
		return nil
	}

	points := make([]Point3, 0, len(r.Line)+2)
	points = append(points, src)
	points = append(points, r.Line...)
	points = append(points, dest)
	return points
}

func (n *navMeshWidget) clearObjects(objects *[]fyne.CanvasObject) {
	for _, obj := range *objects {
		n.container.Remove(obj)
	}
	*objects = nil
}

func insideTriangle(pt, v1, v2, v3 Point3) bool {
	b1 := sign(pt, v1, v2) <= 0
	b2 := sign(pt, v2, v3) <= 0
	b3 := sign(pt, v3, v1) <= 0
	return (b1 == b2) && (b2 == b3)
}

func sign(p1, p2, p3 Point3) float32 {
	return (p1.X-p3.X)*(p2.Y-p3.Y) - (p2.X-p3.X)*(p1.Y-p3.Y)
}
