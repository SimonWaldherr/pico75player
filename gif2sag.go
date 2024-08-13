package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"

	"./imgcolor"
)

// SAGHeader repräsentiert den Header der SAG-Datei.
type SAGHeader struct {
	Signature    [3]byte   // "SAG"
	Version      byte      // Version 1.0 = 0x01
	Width        uint16    // Breite des Bildes in Pixeln
	Height       uint16    // Höhe des Bildes in Pixeln
	FrameCount   uint16    // Anzahl der Frames
	FrameDelay   uint16    // Dauer jedes Frames in Millisekunden
	ColorPalette [768]byte // Globale Farbpalette (256 Farben, je 3 Bytes RGB)
}

// ImageLoader ist eine Schnittstelle zum Laden und Verarbeiten von animierten Bildformaten.
type ImageLoader interface {
	Load(filename string) ([]*image.Paletted, []int, error)
}

// GIFLoader lädt GIF-Bilder.
type GIFLoader struct{}

func (g GIFLoader) Load(filename string) ([]*image.Paletted, []int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	gifImage, err := gif.DecodeAll(file)
	if err != nil {
		return nil, nil, err
	}

	return gifImage.Image, gifImage.Delay, nil
}

// TIFFLoader lädt TIFF-Bilder.
type TIFFLoader struct{}

func (t TIFFLoader) Load(filename string) ([]*image.Paletted, []int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	img, err := tiff.Decode(file)
	if err != nil {
		return nil, nil, err
	}

	return singleFrameToPaletted(img), []int{100}, nil // 100 ms als Standard-Delay
}

// WebPLoader lädt WebP-Bilder.
type WebPLoader struct{}

func (w WebPLoader) Load(filename string) ([]*image.Paletted, []int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	img, err := webp.Decode(file)
	if err != nil {
		return nil, nil, err
	}

	return singleFrameToPaletted(img), []int{100}, nil // 100 ms als Standard-Delay
}

// singleFrameToPaletted konvertiert ein Einzelbild in eine Paletted-Version.
func singleFrameToPaletted(img image.Image) []*image.Paletted {
	bounds := img.Bounds()
	palettedImg := image.NewPaletted(bounds, nil)
	draw.FloydSteinberg.Draw(palettedImg, bounds, img, image.Point{})
	return []*image.Paletted{palettedImg}
}

// reduceColors reduziert die Farbpalette eines Bildes auf 256 Farben.
func reduceColors(frames []*image.Paletted) ([]*image.Paletted, []color.Color) {
	// Erstelle ein gemeinsames ColorCount-Map für alle Frames
	colorCount := make(map[color.Color]int)
	for _, frame := range frames {
		imgcolor.CountColorsInImage(frame, colorCount)
	}

	// Extrahiere die häufigsten 256 Farben
	palette := imgcolor.ExtractPalette(colorCount, 256)

	// Konvertiere alle Frames auf die neue Farbpalette
	for i, frame := range frames {
		frames[i] = applyPalette(frame, palette)
	}

	return frames, palette
}

// applyPalette wendet eine Farbpalette auf ein Bild an.
func applyPalette(frame *image.Paletted, palette []color.Color) *image.Paletted {
	bounds := frame.Bounds()
	newFrame := image.NewPaletted(bounds, palette)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := frame.At(x, y)
			index := imgcolor.NearestColorIndex(palette, oldColor)
			newFrame.SetColorIndex(x, y, uint8(index))
		}
	}

	return newFrame
}

// writeSAGFile erstellt die SAG-Datei aus dem übergebenen animierten Bild.
func writeSAGFile(frames []*image.Paletted, delays []int, palette []color.Color, outputFilename string) error {
	width := uint16(frames[0].Bounds().Dx())
	height := uint16(frames[0].Bounds().Dy())
	frameCount := uint16(len(frames))
	frameDelay := uint16(delays[0] * 10) // Konvertiert 1/100s GIF-Delay in Millisekunden

	// Erstellen und Initialisieren des Headers
	var header SAGHeader
	copy(header.Signature[:], "SAG")
	header.Version = 0x01
	header.Width = width
	header.Height = height
	header.FrameCount = frameCount
	header.FrameDelay = frameDelay

	// Speichere die Farbpalette in den Header
	for i, c := range palette {
		r, g, b, _ := c.RGBA()
		header.ColorPalette[i*3] = uint8(r >> 8)
		header.ColorPalette[i*3+1] = uint8(g >> 8)
		header.ColorPalette[i*3+2] = uint8(b >> 8)
	}

	// Datei erstellen
	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Header in die Datei schreiben
	if err := binary.Write(file, binary.BigEndian, header); err != nil {
		return err
	}

	// Frame-Daten in die Datei schreiben
	for i, frame := range frames {
		prevFrame := (*image.Paletted)(nil)
		if i > 0 {
			prevFrame = frames[i-1]
		}
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(width); x += 8 {
				var identicalByte byte = 0
				var pixelBlock []byte

				for bit := 0; bit < 8; bit++ {
					if x+bit >= int(width) {
						break
					}
					currentPixel := frame.ColorIndexAt(x+bit, y)
					if prevFrame != nil && prevFrame.ColorIndexAt(x+bit, y) == currentPixel {
						identicalByte |= 1 << (7 - bit)
					}
					pixelBlock = append(pixelBlock, currentPixel)
				}

				file.Write([]byte{identicalByte})
				file.Write(pixelBlock)
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: gif2sag <input> <output.sag> <format>")
		fmt.Println("Supported formats: gif, tiff, webp")
		os.Exit(1)
	}

	inputFilename := os.Args[1]
	outputFilename := os.Args[2]
	format := os.Args[3]

	var loader ImageLoader

	switch format {
	case "gif":
		loader = GIFLoader{}
	case "tiff":
		loader = TIFFLoader{}
	case "webp":
		loader = WebPLoader{}
	default:
		fmt.Println("Unsupported format:", format)
		os.Exit(1)
	}

	frames, delays, err := loader.Load(inputFilename)
	if err != nil {
		fmt.Println("Error loading image:", err)
		os.Exit(1)
	}

	// Reduziere die Farben der Frames und extrahiere die Palette
	frames, palette := reduceColors(frames)

	// Schreibe die SAG-Datei
	if err := writeSAGFile(frames, delays, palette, outputFilename); err != nil {
		fmt.Println("Error creating SAG file:", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully:", outputFilename)
}
