package imgcolor

import (
	"image"
	"image/color"
	"image/gif"
	"io"
	"sort"
)

// ColorCount stores the count of pixels for a specific color.
type ColorCount struct {
	Color color.Color
	Count int
}

// CountColors counts the number of pixels per color in an image. It supports both static images and animated GIFs.
func CountColors(r io.Reader) (map[color.Color]int, error) {
	// Attempt to decode the image
	img, _, err := image.Decode(r)
	if err != nil {
		// Check if it's an animated GIF
		if gifImage, gifErr := gif.DecodeAll(r); gifErr == nil {
			return CountColorsInGIF(gifImage), nil
		}
		return nil, err
	}

	// For static images
	colorCount := make(map[color.Color]int)
	CountColorsInImage(img, colorCount)
	return colorCount, nil
}

// CountColorsInImage counts the colors in a static image.
func CountColorsInImage(img image.Image, colorCount map[color.Color]int) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			colorCount[c]++
		}
	}
}

// CountColorsInGIF counts the colors in an animated GIF.
func CountColorsInGIF(gifImage *gif.GIF) map[color.Color]int {
	colorCount := make(map[color.Color]int)

	for _, frame := range gifImage.Image {
		CountColorsInImage(frame, colorCount)
	}

	return colorCount
}

// ExtractPalette returns a palette with the most frequent colors.
// If maxColors == -1, it returns all colors.
func ExtractPalette(colorCount map[color.Color]int, maxColors int) []color.Color {
	// Convert the colorCount map to a slice of ColorCount
	colors := make([]ColorCount, 0, len(colorCount))
	for c, count := range colorCount {
		colors = append(colors, ColorCount{Color: c, Count: count})
	}

	// Sort colors by frequency
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].Count > colors[j].Count
	})

	// Determine the number of colors to return
	if maxColors == -1 || maxColors > len(colors) {
		maxColors = len(colors)
	}

	// Create the palette
	palette := make([]color.Color, maxColors)
	for i := 0; i < maxColors; i++ {
		palette[i] = colors[i].Color
	}

	return palette
}

// NearestColorIndex returns the index of the closest matching color in a palette.
func NearestColorIndex(palette []color.Color, targetColor color.Color) int {
	minDist := int(^uint(0) >> 1) // Maximum int value
	minIndex := 0

	for i, p := range palette {
		dist := colorDistanceSquared(targetColor, p)
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}

	return minIndex
}

// colorDistanceSquared calculates the squared distance between two colors.
func colorDistanceSquared(c1, c2 color.Color) int {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Convert the 16-bit values to 8-bit
	r1 >>= 8
	g1 >>= 8
	b1 >>= 8
	r2 >>= 8
	g2 >>= 8
	b2 >>= 8

	// Euclidean distance in RGB color space
	rd := int(r1) - int(r2)
	gd := int(g1) - int(g2)
	bd := int(b1) - int(b2)

	return rd*rd + gd*gd + bd*bd
}
