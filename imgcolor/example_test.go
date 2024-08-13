package imgcolor

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"sort"
)

// sortColors sorts the colors by count descending, then by RGBA components (R, G, B, A) ascending
func sortColors(colors []ColorCount) {
	sort.SliceStable(colors, func(i, j int) bool {
		if colors[i].Count == colors[j].Count {
			r1, g1, b1, a1 := colors[i].Color.RGBA()
			r2, g2, b2, a2 := colors[j].Color.RGBA()
			if r1 != r2 {
				return r1 < r2
			}
			if g1 != g2 {
				return g1 < g2
			}
			if b1 != b2 {
				return b1 < b2
			}
			return a1 < a2
		}
		return colors[i].Count > colors[j].Count
	})
}

// Example demonstrates how to use the imgcolor package to analyze an example GIF file.
func Example() {
	// Open the example.gif file
	file, err := os.Open("example.gif")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Count the colors in the GIF
	colorCount, err := CountColors(file)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the most frequent 256 colors
	palette := ExtractPalette(colorCount, 9)

	// Convert the colorCount map to a slice of ColorCount
	colors := make([]ColorCount, 0, len(colorCount))
	for c, count := range colorCount {
		colors = append(colors, ColorCount{Color: c, Count: count})
	}

	// Sort the colors
	sortColors(colors)

	// Output the palette
	fmt.Println("Palette (256 colors):")
	for i, _ := range colors[:20] {
		c := colors[i].Color
		fmt.Printf("%d: %v %d\n", i, c, colorCount[c])
	}

	// Comment out the nearest color index part if it is not needed for the test
	exampleColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red as an example
	index := NearestColorIndex(palette, exampleColor)
	fmt.Printf("The nearest color index for red is: %d\n", index)

	// Output:
	// Palette (256 colors):
	// 0: {0 0 0 255} 3531
	// 1: {198 0 60 255} 11
	// 2: {211 122 122 255} 10
	// 3: {0 145 139 255} 9
	// 4: {165 0 50 255} 9
	// 5: {193 0 58 255} 9
	// 6: {0 60 58 255} 8
	// 7: {177 0 54 255} 8
	// 8: {0 81 77 255} 7
	// 9: {0 170 163 255} 7
	// 10: {61 35 35 255} 6
	// 11: {79 0 24 255} 6
	// 12: {120 0 36 255} 6
	// 13: {161 0 49 255} 6
	// 14: {214 124 124 255} 6
	// 15: {232 0 70 255} 6
	// 16: {240 138 138 255} 6
	// 17: {248 143 143 255} 6
	// 18: {0 27 25 255} 5
	// 19: {0 68 66 255} 5
	// The nearest color index for red is: 1
}
