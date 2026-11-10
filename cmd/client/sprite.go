package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type spriteStore struct {
	images   map[string]image.Image
	rendered map[string]string
}

func newSpriteStore() *spriteStore {
	return &spriteStore{images: map[string]image.Image{}, rendered: map[string]string{}}
}

func (s *spriteStore) get(path string, maxW, maxH int) string {
	key := fmt.Sprintf("%s|%d|%d", path, maxW, maxH)
	if r, ok := s.rendered[key]; ok {
		return r
	}
	img, ok := s.images[path]
	if !ok {
		img = loadImage(path)
		s.images[path] = img
	}
	r := ""
	if img != nil {
		r = renderHalfBlocks(img, maxW, maxH)
	}
	s.rendered[key] = r
	return r
}

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return img
}

func opaqueBounds(img image.Image) image.Rectangle {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
				minX, maxX = min(minX, x), max(maxX, x)
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
	}
	if maxX < minX {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

type pixel struct {
	color lipgloss.Color
	ok    bool
}

func renderHalfBlocks(img image.Image, maxW, maxH int) string {
	src := opaqueBounds(img)
	if src.Empty() || maxW < 1 || maxH < 1 {
		return ""
	}
	cw, ch := float64(src.Dx()), float64(src.Dy())
	scale := math.Min(float64(maxW)/cw, float64(2*maxH)/ch)
	if scale > 1 {
		scale = math.Floor(scale)
	}
	outW := max(1, int(cw*scale))
	outH := max(2, int(ch*scale))
	if outH%2 == 1 {
		outH++
	}

	sample := func(x, y int) pixel {
		x0 := src.Min.X + int(float64(x)/scale)
		y0 := src.Min.Y + int(float64(y)/scale)
		x1 := min(src.Max.X, max(x0+1, src.Min.X+int(float64(x+1)/scale)))
		y1 := min(src.Max.Y, max(y0+1, src.Min.Y+int(float64(y+1)/scale)))
		var r, g, b, a, n float64
		for sy := y0; sy < y1; sy++ {
			for sx := x0; sx < x1; sx++ {
				pr, pg, pb, pa := img.At(sx, sy).RGBA()
				r, g, b, a = r+float64(pr), g+float64(pg), b+float64(pb), a+float64(pa)
				n++
			}
		}
		if n == 0 || a/n < 0.5*0xffff {
			return pixel{}
		}

		hex := fmt.Sprintf("#%02x%02x%02x", uint8(r/a*255), uint8(g/a*255), uint8(b/a*255))
		return pixel{color: lipgloss.Color(hex), ok: true}
	}

	var sb strings.Builder
	for y := 0; y < outH; y += 2 {
		if y > 0 {
			sb.WriteByte('\n')
		}
		for x := 0; x < outW; x++ {
			top, bot := sample(x, y), sample(x, y+1)
			switch {
			case top.ok && bot.ok:
				sb.WriteString(lipgloss.NewStyle().Foreground(top.color).Background(bot.color).Render("▀"))
			case top.ok:
				sb.WriteString(lipgloss.NewStyle().Foreground(top.color).Render("▀"))
			case bot.ok:
				sb.WriteString(lipgloss.NewStyle().Foreground(bot.color).Render("▄"))
			default:
				sb.WriteByte(' ')
			}
		}
	}
	return sb.String()
}
