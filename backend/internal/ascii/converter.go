package ascii

import (
	"io"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

const charSet = " .:-=+*#%@"

type Converter struct {
	Options Options
}

type Options struct {
	TargetWidth int
	Mode string
}

func NewConverter(opts Options) *Converter {
	if opts.TargetWidth == 0 {
		opts.TargetWidth = 100
	}
	return &Converter{Options: opts}
}

func (c *Converter) Convert(img image.Image) string {
	bounds := img.Bounds()
	widthIn := bounds.Dx()
	heightIn := bounds.Dy()
	targetHeight := int(float64(c.Options.TargetWidth) * float64(heightIn) / float64(widthIn) * 0.55)
	if targetHeight == 0 && heightIn > 0 {
		targetHeight = 1
	}

	var builder strings.Builder
	builder.Grow((c.Options.TargetWidth + 1) * targetHeight)

	if pImg, ok := img.(*image.Paletted); ok {
		lut := make([]string, len(pImg.Palette))
		for i, col := range pImg.Palette {
			lut[i] = c.pixelToChar(col)
		}

		for y := range targetHeight {
			srcY := y * heightIn / targetHeight
			for x := range c.Options.TargetWidth {
				srcX := x * widthIn / c.Options.TargetWidth

				colorIndex := pImg.ColorIndexAt(srcX, srcY)
				builder.WriteString(lut[colorIndex])
			}
			builder.WriteByte('\n')
		}
		return builder.String()
	}

	for y := range targetHeight {
		srcY := y * heightIn / targetHeight
		for x := range c.Options.TargetWidth {
			srcX := x * widthIn / c.Options.TargetWidth

			p := img.At(srcX, srcY)
			builder.WriteString(c.pixelToChar(p))
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (c *Converter) pixelToChar(p color.Color) string {
	r, g, b, _ := p.RGBA()

	lum := (r*19595 + g*38470 + b*7471) >> 16

	index := int(lum) * (len(charSet) - 1) / 65535
	return string(charSet[index])
}

func Decode(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

