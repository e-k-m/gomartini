package martini

import (
	"fmt"
	"math"
)

// Martini after instantiation using New holds constants needed for the
// CreateTile and GetMesh steps.
type Martini struct {
	gridSize           int
	numTriangles       int
	numParentTriangles int
	indices            []int
	coords             []int
}

// New creates a Martini.
//
// # Parameters
//
//   - gridSize: the grid size to use when generating the mesh. Must be 2^k+1.
//     If your source heightmap is 256x256 pixels, use grid_size=257 and backfill
//     the border pixels.
//
// # Returns
//
// Returns a Martini on which you can call CreateTile.
func New(gridSize int) Martini {
	tileSize := gridSize - 1
	if tileSize&(tileSize-1) != 0 {
		panic(
			fmt.Sprintf(
				"martini: expected grid size to be 2^n+1, got %d.",
				gridSize,
			),
		)
	}

	numTriangles := tileSize*tileSize*2 - 2
	numParentTriangles := numTriangles - tileSize*tileSize

	indices := make([]int, gridSize*gridSize)

	coords := make([]int, numTriangles*4)

	for i := 0; i < numTriangles; i++ {
		id := i + 2
		ax, ay, bx, by, cx, cy := 0, 0, 0, 0, 0, 0
		if id&1 != 0 {
			bx, by, cx = tileSize, tileSize, tileSize
		} else {
			ax, ay, cy = tileSize, tileSize, tileSize
		}
		id >>= 1
		for id > 1 {
			mx := (ax + bx) >> 1
			my := (ay + by) >> 1

			if id&1 != 0 {
				bx, by = ax, ay
				ax, ay = cx, cy
			} else {
				ax, ay = bx, by
				bx, by = cx, cy
			}
			cx = mx
			cy = my
			id >>= 1
		}

		k := i * 4
		coords[k+0] = ax
		coords[k+1] = ay
		coords[k+2] = bx
		coords[k+3] = by
	}

	return Martini{
		gridSize:           gridSize,
		numTriangles:       numTriangles,
		numParentTriangles: numParentTriangles,
		indices:            indices,
		coords:             coords,
	}

}

// CreateTile generates RTIN hierarchy from terrain data.
//
// # Parameters
//
//   - terrain: representing the input heightmap. The array must be flattened, of
//     shape 2^k+1 * 2^k+1.
//
// # Returns
//
// Returns a Tile on which you can call GetMesh.
func (m *Martini) CreateTile(terrain []float64) Tile {
	return NewTile(terrain, m)
}

// Tile after instantiation using NewTile holds constants needed for the
// GetMesh steps.
type Tile struct {
	terrain []float64
	martini *Martini
	errors  []float64
}

// NewTile creates a Tile.
//
// # Parameters
//
//   - terrain: representing the input heightmap. The array must be flattened, of
//     shape 2^k+1 * 2^k+1.
//
//   - martini: a Martini.
//
// # Returns
//
// Returns a Tile on which you can call GetMesh.
func NewTile(terrain []float64, martini *Martini) Tile {
	size := martini.gridSize
	if len(terrain) != size*size {
		panic(
			fmt.Sprintf(
				"martini: expected terrain data of length %d, got %d",
				size*size,
				len(terrain),
			),
		)
	}
	t := Tile{
		terrain: terrain,
		martini: martini,
		errors:  make([]float64, len(terrain)),
	}
	t.update()
	return t
}

// GetMesh gets a mesh for a given max error.
//
// # Parameters
//
//   - maxError: the maximum vertical error for each triangle in the output mesh.
//     For example if the units of the input heightmap is meters, using
//     maxError=5 would mean that the mesh is continually refined until every
//     triangle approximates the surface of the heightmap within 5 meters.
//
// # Returns
//
// Returns slices of vertices and triangles. Vertices represents the interleaved
// 2D coordinates of each vertex, e.g. [x0, y0, x1, y1, ...]. Triangles represents
// indices within the vertices array. So [0, 1, 3, ...] would use the first,
// second, and fourth vertices within the vertices array as a single triangle.
func (t *Tile) GetMesh(maxError float64) (vertices []int, triangles []int) {
	numVertices := 0
	numTriangles := 0

	max := t.martini.gridSize - 1

	for i := range t.martini.indices {
		t.martini.indices[i] = 0
	}

	var countElements func(ax, ay, bx, by, cx, cy int)
	countElements = func(ax, ay, bx, by, cx, cy int) {
		mx := (ax + bx) >> 1
		my := (ay + by) >> 1

		if math.Abs(float64(ax-cx))+math.Abs(float64(ay-cy)) > 1.0 &&
			t.errors[my*t.martini.gridSize+mx] > maxError {
			countElements(cx, cy, ax, ay, mx, my)
			countElements(bx, by, cx, cy, mx, my)
		} else {
			if t.martini.indices[ay*t.martini.gridSize+ax] == 0 {
				numVertices++
				t.martini.indices[ay*t.martini.gridSize+ax] = numVertices
			}
			if t.martini.indices[by*t.martini.gridSize+bx] == 0 {
				numVertices++
				t.martini.indices[by*t.martini.gridSize+bx] = numVertices
			}
			if t.martini.indices[cy*t.martini.gridSize+cx] == 0 {
				numVertices++
				t.martini.indices[cy*t.martini.gridSize+cx] = numVertices
			}
			numTriangles++
		}

	}
	countElements(0, 0, max, max, max, 0)
	countElements(max, max, 0, 0, 0, max)

	vertices = make([]int, numVertices*2)
	triangles = make([]int, numTriangles*3)
	triIndex := 0

	var processTriangle func(ax, ay, bx, by, cx, cy int)
	processTriangle = func(ax, ay, bx, by, cx, cy int) {
		mx := (ax + bx) >> 1
		my := (ay + by) >> 1

		if math.Abs(float64(ax-cx))+math.Abs(float64(ay-cy)) > 1 &&
			t.errors[my*t.martini.gridSize+mx] > maxError {
			processTriangle(cx, cy, ax, ay, mx, my)
			processTriangle(bx, by, cx, cy, mx, my)
		} else {
			a := t.martini.indices[ay*t.martini.gridSize+ax] - 1
			b := t.martini.indices[by*t.martini.gridSize+bx] - 1
			c := t.martini.indices[cy*t.martini.gridSize+cx] - 1

			vertices[2*a] = ax
			vertices[2*a+1] = ay

			vertices[2*b] = bx
			vertices[2*b+1] = by

			vertices[2*c] = cx
			vertices[2*c+1] = cy

			triangles[triIndex] = a
			triIndex++
			triangles[triIndex] = b
			triIndex++
			triangles[triIndex] = c
			triIndex++

		}

	}

	processTriangle(0, 0, max, max, max, 0)
	processTriangle(max, max, 0, 0, 0, max)

	return vertices, triangles
}

func (t *Tile) update() {
	for i := t.martini.numTriangles - 1; i >= 0; i-- {
		k := i * 4
		ax := t.martini.coords[k+0]
		ay := t.martini.coords[k+1]
		bx := t.martini.coords[k+2]
		by := t.martini.coords[k+3]
		mx := (ax + bx) >> 1
		my := (ay + by) >> 1
		cx := mx + my - ay
		cy := my + ax - mx

		interpolatedHeight := (t.terrain[ay*t.martini.gridSize+ax] + t.terrain[by*t.martini.gridSize+bx]) / 2
		middleIndex := my*t.martini.gridSize + mx

		middleError := math.Abs(
			float64(interpolatedHeight - t.terrain[middleIndex]),
		)

		t.errors[middleIndex] = float64(
			math.Max(float64(t.errors[middleIndex]), float64(middleError)),
		)

		if i < t.martini.numParentTriangles {
			leftChildIndex := ((ay+cy)>>1)*t.martini.gridSize + ((ax + cx) >> 1)
			rightChildIndex := ((by+cy)>>1)*t.martini.gridSize + ((bx + cx) >> 1)

			t.errors[middleIndex] = float64(
				math.Max(
					math.Max(
						float64(t.errors[middleIndex]),
						float64(t.errors[leftChildIndex]),
					),
					float64(t.errors[rightChildIndex]),
				),
			)
		}
	}
}
