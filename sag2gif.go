package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
)

// SAGHeader represents the header of the SAG file.
type SAGHeader struct {
	Signature    [3]byte
	Version      byte
	Width        uint16
	Height       uint16
	FrameCount   uint16
	FrameDelay   uint16
	ColorPalette [768]byte
}

// readSAGFile reads a SAG file and returns the frames and the delays between them.
func readSAGFile(filename string) ([]*image.Paletted, []int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	header, err := readSAGHeader(file)
	if err != nil {
		return nil, nil, err
	}

	frames, delays, err := readSAGFrames(file, header)
	if err != nil {
		return nil, nil, err
	}

	return frames, delays, nil
}

// readSAGHeader reads the SAG header from the file.
func readSAGHeader(file *os.File) (SAGHeader, error) {
	var header SAGHeader
	if err := binary.Read(file, binary.BigEndian, &header); err != nil {
		return header, err
	}
	return header, nil
}

// readSAGFrames reads the frames and delays from the SAG file.
func readSAGFrames(file *os.File, header SAGHeader) ([]*image.Paletted, []int, error) {
	palette := extractPalette(header)
	width, height := int(header.Width), int(header.Height)
	frameCount, frameDelay := int(header.FrameCount), int(header.FrameDelay)

	frames := make([]*image.Paletted, frameCount)
	delays := make([]int, frameCount)

	for i := 0; i < frameCount; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, width, height), palette)

		for y := 0; y < height; y++ {
			for x := 0; x < width; x += 8 {
				skipIdenticalByte(file)

				pixelBlock, err := readPixelBlock(file, width, x)
				if err != nil {
					return nil, nil, err
				}

				applyPixelBlock(frame, pixelBlock, x, y, width)
			}
		}

		frames[i] = frame
		delays[i] = frameDelay / 10 // Convert back to 1/100th of a second for GIF
	}

	return frames, delays, nil
}

// extractPalette creates a color palette from the SAG header.
func extractPalette(header SAGHeader) color.Palette {
	palette := make([]color.Color, 256)
	for i := 0; i < 256; i++ {
		r, g, b := header.ColorPalette[i*3], header.ColorPalette[i*3+1], header.ColorPalette[i*3+2]
		palette[i] = color.RGBA{R: r, G: g, B: b, A: 0xff}
	}
	return palette
}

// skipIdenticalByte skips the identical byte in the SAG file.
func skipIdenticalByte(file *os.File) {
	file.Read(make([]byte, 1))
}

// readPixelBlock reads the next 8 pixels from the SAG file.
func readPixelBlock(file *os.File, width, x int) ([]byte, error) {
	pixelBlock := make([]byte, 8)
	if x+8 > width {
		pixelBlock = make([]byte, width-x)
	}
	_, err := file.Read(pixelBlock)
	return pixelBlock, err
}

// applyPixelBlock applies a block of pixels to a frame.
func applyPixelBlock(frame *image.Paletted, pixelBlock []byte, x, y, width int) {
	for bit := 0; bit < len(pixelBlock); bit++ {
		if x+bit < width {
			frame.SetColorIndex(x+bit, y, pixelBlock[bit])
		}
	}
}

// writeGIFFile writes the frames and delays as a GIF file with infinite looping.
func writeGIFFile(frames []*image.Paletted, delays []int, outputFilename string) error {
	outGif := &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0, // Infinite loop
	}

	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	return gif.EncodeAll(file, outGif)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: sag2gif <input.sag> <output.gif>")
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	outputFilename := os.Args[2]

	frames, delays, err := readSAGFile(inputFilename)
	if err != nil {
		fmt.Println("Error reading SAG file:", err)
		os.Exit(1)
	}

	if err := writeGIFFile(frames, delays, outputFilename); err != nil {
		fmt.Println("Error writing GIF file:", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully:", outputFilename)
}
