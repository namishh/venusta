package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/esimov/stackblur-go"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TYPES

type model struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

type colors struct {
	base8   []string
	pallete []string
}

type ColorSlice []color.RGBA

// SORT FUNCTIONS
func (p ColorSlice) Len() int      { return len(p) }
func (p ColorSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ColorSlice) Less(i, j int) bool {
	// Compare red components
	if p[i].R != p[j].R {
		return p[i].R < p[j].R
	}

	// Compare green components
	if p[i].G != p[j].G {
		return p[i].G < p[j].G
	}

	// Compare blue components
	return p[i].B < p[j].B
}

// PATH FUNCTIONS

func getImage(imagePath string) (img image.Image, err error) {
	file, err := os.Open(imagePath)
	if err != nil {
		err = fmt.Errorf("Failure: %s", err)
		return
	}

	defer file.Close()

	mimeType := path.Ext(imagePath)
	mimeType = strings.TrimPrefix(mimeType, ".")

	switch mimeType {
	case "jpeg", "jpg":
		img, err = jpeg.Decode(file)

	case "png":
		img, err = png.Decode(file)
	}

	return
}

// IMAGE FUNCTIONS

func sliceImage(img image.Image, grid int) (tiles []image.Image) {
	tiles = make([]image.Image, 0, grid*grid)

	if cap(tiles) == 0 {
		return
	}

	shape := img.Bounds()

	fheight := float64(shape.Max.Y / int(grid))
	fwidth := float64(shape.Max.X / int(grid))

	height := int(math.Ceil(fheight))
	width := int(math.Ceil(fwidth))

	for y := shape.Min.Y; y+height <= shape.Max.Y; y += height {

		for x := shape.Min.X; x+width <= shape.Max.X; x += width {

			tile := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+width, y+height))

			tiles = append(tiles, tile)
		}
	}

	return
}

func getPallete(img image.Image) (cols []color.RGBA, hexColors []string) {

	blurred_img, _ := stackblur.Process(img, 30)

	tiles := sliceImage(blurred_img, 4)

	cols = make([]color.RGBA, 0)
	hexColors = make([]string, 0)
	for _, tile := range tiles {
		colors := make(map[color.RGBA]int)
		bounds := tile.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				col := tile.At(x, y)
				r, g, b, a := col.RGBA()
				rgbaCol := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
				colors[rgbaCol]++

			}
		}
		var mostFrequentColor color.RGBA
		maxCount := 0
		for col, count := range colors {
			if count > maxCount {
				maxCount = count
				mostFrequentColor = col
			}
		}
		hexColor := fmt.Sprintf("#%02x%02x%02x", mostFrequentColor.R, mostFrequentColor.G, mostFrequentColor.B)

		hexColors = append(hexColors, hexColor)
		cols = append(cols, mostFrequentColor)

	}

	return
}

func getBase8(img image.Image) (colors []color.RGBA, hexColors []string) {
	darkestColor := color.RGBA{255, 255, 255, 255} // Initialize with the brightest color
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			col := img.At(x, y)
			r, g, b, _ := col.RGBA()
			rgbaCol := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}

			// Calculate luminance
			luminance := 0.2126*float64(rgbaCol.R) + 0.7152*float64(rgbaCol.G) + 0.0722*float64(rgbaCol.B)

			// Update darkestColor if this color has lower luminance
			if luminance < 0.2126*float64(darkestColor.R)+0.7152*float64(darkestColor.G)+0.0722*float64(darkestColor.B) {
				darkestColor = rgbaCol
			}
		}
	}

	n := 8
	colors = make([]color.RGBA, 0)
	hexColors = make([]string, 0)
	for i := 0; i < n; i++ {
		colors = append(colors, darkestColor)

		// Make the color 12.5% lighter (12.5 * 8 == 100)
		darkestColor.R = uint8(float64(darkestColor.R) + 0.125*255)
		darkestColor.G = uint8(float64(darkestColor.G) + 0.125*255)
		darkestColor.B = uint8(float64(darkestColor.B) + 0.125*255)

		// Make sure the color components don't exceed 255
		if darkestColor.R > 255 {
			darkestColor.R = 255
		}
		if darkestColor.G > 255 {
			darkestColor.G = 255
		}
		if darkestColor.B > 255 {
			darkestColor.B = 255
		}

	}

	for _, col := range colors {
		hexColor := fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
		hexColors = append(hexColors, hexColor)
	}

	return
}

func sortHexColors(hexColors []string) []string {
	// Convert the hex colors to color.RGBA values
	colors := make([]color.RGBA, len(hexColors))
	for i, hex := range hexColors {
		r, _ := strconv.ParseUint(hex[1:3], 16, 8)
		g, _ := strconv.ParseUint(hex[3:5], 16, 8)
		b, _ := strconv.ParseUint(hex[5:7], 16, 8)
		colors[i] = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}

	// Sort the colors
	sort.Sort(ColorSlice(colors))

	// Convert the sorted colors back to hex codes
	sortedHexColors := make([]string, len(colors))
	for i, c := range colors {
		sortedHexColors[i] = fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}

	return sortedHexColors
}

// BUBBLES FUNCTIONS

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
		return m, tea.Quit
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}

// MAIN FUNCTION

func main() {

	fp := filepicker.New()
	fp.AllowedTypes = []string{".jpg", ".jpeg", ".png"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	m := model{
		filepicker: fp,
	}
	tm, _ := tea.NewProgram(&m).Run()
	mm := tm.(model)

	fmt.Println("\n  You selected: " + m.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")

	img, _ := getImage(mm.selectedFile)
	_, hexbases := getBase8(img)
	_, hexpal := getPallete(img)
	sortedHexColors := sortHexColors(hexpal)
	for i, j := 0, len(sortedHexColors)-1; i < j; i, j = i+1, j-1 {
		sortedHexColors[i], sortedHexColors[j] = sortedHexColors[j], sortedHexColors[i]
	}

	fmt.Println("Background to foreground")
	fmt.Println(hexbases)

	fmt.Println("Prominent Colors")
	fmt.Println(sortedHexColors)
}
