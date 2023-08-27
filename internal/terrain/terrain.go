package terrain

import (
	"image"
	_ "image/png"
	"os"
)

type Terrain struct {
	Terrain  []float64
	GridSize int
}

func Read(path string) (Terrain, error) {
	img, err := readImage(path)
	if err != nil {
		return Terrain{}, err
	}
	return Terrain{
		Terrain:  terrain(img),
		GridSize: img.Bounds().Dx() + 1,
	}, nil
}

func readImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func terrain(img image.Image) []float64 {
	gridSize := img.Bounds().Dx() + 1
	terrain := make([]float64, gridSize*gridSize)
	tileSize := img.Bounds().Dx()

	for y := 0; y < tileSize; y++ {
		for x := 0; x < tileSize; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r /= 257
			g /= 257
			b /= 257
			terrain[y*gridSize+x] = (float64(r)*256.0*256.0+float64(g)*256.0+float64(b))/10.0 - 10000.0
		}
	}

	for x := 0; x < gridSize-1; x++ {
		terrain[gridSize*(gridSize-1)+x] = terrain[gridSize*(gridSize-2)+x]
	}

	for y := 0; y < gridSize; y++ {
		terrain[gridSize*y+gridSize-1] = terrain[gridSize*y+gridSize-2]
	}

	return terrain
}
