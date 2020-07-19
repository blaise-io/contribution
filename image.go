package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"

	"github.com/esimov/colorquant"
	"github.com/nfnt/resize"
)

// ColorMappedImage represents a gray image with reduced color palette,
// and an indexed map of quantized colors in the image, sorted from
// light to dark.
type ColorMappedImage struct {
	image    *image.Gray
	colorMap map[int]int
}

// GithubPalette is the palette of GitHub contribution graph.
var GithubPalette = color.Palette([]color.Color{
	color.RGBA{235, 237, 240, 255}, // No activity
	color.RGBA{155, 233, 168, 255}, // Low activity
	color.RGBA{64, 196, 99, 255},
	color.RGBA{48, 161, 78, 255},
	color.RGBA{33, 110, 57, 255}, // High activity
})

// ReadImage reads the image file at path and decodes it to PNG.
func ReadImage(imageFile string) (image.Image, error) {
	file, err := os.Open(imageFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	img, _, err := image.Decode(file)
	return img, err
}

// ResizeImage resizes the image to fit the GitHub contribution graph bounds.
func ResizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	if bounds.Max.Y > 7 {
		w := uint(bounds.Max.X / (bounds.Max.Y / 7))
		img = resize.Resize(w, 7, img, resize.NearestNeighbor)
	}
	return img
}

// ToGithubPalette converts gray colors in the quantified image to the GitHub
// activity palette.
func ToGithubPalette(img image.Image) image.Image {
	colorMapped := getColorMappedImage(img)

	bounds := img.Bounds()
	githubImg := image.NewRGBA(bounds)

	// Replace colors
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			gray := colorMapped.image.GrayAt(x, y)
			paletteIndex := colorMapped.colorMap[int(gray.Y)]
			githubImg.Set(x, y, GithubPalette[paletteIndex])
		}
	}

	return githubImg
}

// PreviewResult show a preview of the image in shell using block characters.
func PreviewResult(img image.Image) {
	colorMapped := getColorMappedImage(img)
	bounds := img.Bounds()

	// Replace colors
	for y := 0; y < bounds.Max.Y; y++ {
		fmt.Println()
		for x := 0; x < bounds.Max.X; x++ {
			gray := colorMapped.image.GrayAt(x, y)
			idx := colorMapped.colorMap[int(gray.Y)]
			fmt.Print(PixelChr(idx))
		}
	}
	fmt.Println()
}

// SavePNG stores the image as a PNG file.
func SavePNG(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// getImageWithColorMap returns the image with reduced colors in grayscale and
// includes its colormap.
func getColorMappedImage(img image.Image) ColorMappedImage {
	bounds := img.Bounds()

	// Drawing image onto a white surface to remove transparency
	opaqueImg := image.NewRGBA(bounds)
	draw.Draw(opaqueImg, bounds, image.White, image.Point{}, draw.Src)
	draw.Draw(opaqueImg, bounds, img, image.Point{}, draw.Over)

	// Reduce palette to number of possible colors in GitHub graph
	quantizedImg := image.NewGray(bounds)
	numColors := len(GithubPalette)
	colorquant.NoDither.Quantize(opaqueImg, quantizedImg, numColors, true, true)

	// Create map of all unique grays in the quantized image
	// Mapped from gray.Y to GithubPalette index (populated later)
	colorMap := map[int]int{}
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			gray := quantizedImg.GrayAt(x, y)
			colorMap[int(gray.Y)] = 0
		}
	}

	// Create slice with grays sorted asc, darkest first
	colorsInMap := []int{}
	for k := range colorMap {
		colorsInMap = append(colorsInMap, k)
	}
	sort.Ints(colorsInMap)

	// Populate color map values with order of grays
	for index, k := range colorsInMap {
		colorMap[k] = len(colorsInMap) - 1 - index
	}

	return ColorMappedImage{
		image:    quantizedImg,
		colorMap: colorMap,
	}
}
