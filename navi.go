package navmesh

import (
	"errors"
	_ "fmt"
	. "github.com/spate/vectormath"
)

var (
	ERROR_TRIANGLELIST_ILLEGAL = errors.New("triangle list illegal")
)

type TriangleList struct {
	Vertices  []Point3
	Triangles [][3]int32 // triangles
}

type BorderList struct {
	Indices []int32 // 2pt as border
}

type Path struct {
	Line []Point3
}

type NavMesh struct{}

func (nm *NavMesh) Route(list TriangleList, start, end *Point3) (*Path, error) {
	r := Path{}
	// 计算临边
	border := nm.create_border(list.Triangles)
	// 目标点
	vertices := append(list.Vertices, *end)
	border = append(border, int32(len(vertices)), int32(len(vertices)))

	// 第一个可视区域
	line_start := start
	last_vis_left, last_vis_right, last_p_left, last_p_right := nm.update_vis(start, vertices, border, 0, 1)
	var res Vector3
	for k := 2; k <= len(border)-2; k += 2 {
		cur_vis_left, cur_vis_right, p_left, p_right := nm.update_vis(line_start, vertices, border, k, k+1)
		V3Cross(&res, last_vis_left, cur_vis_right)
		if res.Z > 0 { // 左拐点
			line_start = &vertices[border[last_p_left]]
			r.Line = append(r.Line, *line_start)
			// 找到一条不共点的边作为可视区域
			i := 2 * (last_p_left/2 + 1)
			for ; i <= len(border)-2; i += 2 {
				if border[last_p_left] != border[i] && border[last_p_left] != border[i+1] {
					last_vis_left, last_vis_right, last_p_left, last_p_right = nm.update_vis(line_start, vertices, border, i, i+1)
					break
				}
			}

			k = i
			continue
		}

		V3Cross(&res, last_vis_right, cur_vis_left)
		if res.Z < 0 { // 右拐点
			line_start = &vertices[border[last_p_right]]
			r.Line = append(r.Line, *line_start)
			// 找到一条不共点的边
			i := 2 * (last_p_right/2 + 1)
			for ; i <= len(border)-2; i += 2 {
				if border[last_p_right] != border[i] && border[last_p_right] != border[i+1] {
					last_vis_left, last_vis_right, last_p_left, last_p_right = nm.update_vis(line_start, vertices, border, i, i+1)
					break
				}
			}

			k = i
			continue
		}

		V3Cross(&res, last_vis_left, cur_vis_left)
		if res.Z < 0 {
			last_vis_left = cur_vis_left
			last_p_left = p_left
		}

		V3Cross(&res, last_vis_right, cur_vis_right)
		if res.Z > 0 {
			last_vis_right = cur_vis_right
			last_p_right = p_right
		}
	}

	return &r, nil
}

func (nm *NavMesh) create_border(list [][3]int32) []int32 {
	var border []int32
	for k := 0; k < len(list)-1; k++ {
		for _, i := range list[k] {
			for _, j := range list[k+1] {
				if i == j {
					border = append(border, i)
				}
			}
		}
	}
	return border
}

func (nm *NavMesh) update_vis(v0 *Point3, vertices []Point3, indices []int32, i1, i2 int) (l, r *Vector3, left, right int) {
	var left_vec, right_vec, res Vector3
	P3Sub(&left_vec, &vertices[indices[i1]], v0)
	P3Sub(&right_vec, &vertices[indices[i2]], v0)
	V3Cross(&res, &left_vec, &right_vec)
	if res.Z > 0 {
		return &right_vec, &left_vec, i2, i1
	} else {
		return &left_vec, &right_vec, i1, i2
	}
}
