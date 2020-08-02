package porygion

import (
	"image"
	"image/color"
)

// Standard colors for various properties on the region map.
var (
	colorWater0      = color.RGBA{152, 208, 248, 255}
	colorWater1      = color.RGBA{160, 176, 248, 255}
	colorLand0       = color.RGBA{0, 112, 0, 255}
	colorLand1       = color.RGBA{56, 168, 8, 255}
	colorLand2       = color.RGBA{96, 208, 0, 255}
	colorLand3       = color.RGBA{168, 232, 48, 255}
	colorLand4       = color.RGBA{208, 248, 120, 255}
	colorRouteWater0 = color.RGBA{72, 152, 224, 255}
	colorRouteWater1 = color.RGBA{40, 128, 224, 255}
	colorRouteLand0  = color.RGBA{224, 160, 0, 255}
	colorRouteLand1  = color.RGBA{232, 184, 56, 255}
	colorRouteLand2  = color.RGBA{240, 208, 80, 255}
	colorRouteLand3  = color.RGBA{232, 224, 112, 255}
	colorRouteLand4  = color.RGBA{232, 224, 168, 255}
)

var routeConversionColors = map[color.Color]color.RGBA{
	colorWater0: colorRouteWater0,
	colorWater1: colorRouteWater1,
	colorLand0:  colorRouteLand0,
	colorLand1:  colorRouteLand1,
	colorLand2:  colorRouteLand2,
	colorLand3:  colorRouteLand3,
	colorLand4:  colorRouteLand4,
}

func renderRegionMapImage(elevations [][]float64, cities []Tile, routes []Tile) image.Image {
	width := len(elevations)
	height := len(elevations[0])
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			c := getColorForElevation(elevations[i][j], j)
			img.SetRGBA(i, j, c)
		}
	}
	for _, route := range routes {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				x := route.X*8 + i
				y := route.Y*8 + j
				c := routeConversionColors[img.At(x, y)]
				img.SetRGBA(x, y, c)
			}
		}
	}
	for _, city := range cities {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				x := city.X*8 + i
				y := city.Y*8 + j
				img.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}
	return img
}

func getColorForElevation(elevation float64, y int) color.RGBA {
	if elevation > 0 {
		switch {
		case elevation > 1.10:
			return colorLand4
		case elevation > 0.85:
			return colorLand3
		case elevation > 0.60:
			return colorLand2
		case elevation > 0.35:
			return colorLand1
		default:
			return colorLand0
		}
	}

	// The water alternates blue hues each row.
	if y%2 == 0 {
		return colorWater0
	}
	return colorWater1
}
