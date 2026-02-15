package ascii

import (
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strconv"
	"strings"
)

const charSet = ".:-=+*>#%@"

type Converter struct {
	Options Options
}

type Options struct {
	TargetWidth int
	TargetHeight int
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

	var targetWidth, targetHeight int

	if c.Options.TargetWidth > 0 && c.Options.TargetHeight > 0 {
		targetWidth = c.Options.TargetWidth
		targetHeight = c.Options.TargetHeight
	} else if c.Options.TargetHeight > 0 {
		targetHeight = c.Options.TargetHeight
		targetWidth = int(float64(targetHeight) * (float64(widthIn) / float64(heightIn)) / 0.55)
	} else {
		targetWidth = c.Options.TargetWidth
		if targetWidth == 0 { targetWidth = 100 }
		targetHeight = int(float64(targetWidth) * (float64(heightIn) / float64(widthIn)) / 0.55)
	}
	if targetHeight == 0 && heightIn > 0 { targetHeight = 1	}
	if targetWidth == 0 && widthIn > 0 { targetWidth = 1 }

	var builder strings.Builder

	if c.Options.Mode == "ansi" {
		builder.Grow((c.Options.TargetWidth * 25) * targetHeight)
	} else {
		builder.Grow((c.Options.TargetWidth + 1) * targetHeight)
	}

	if pImg, ok := img.(*image.Paletted); ok {
		lut := make([]string, len(pImg.Palette))
		for i, col := range pImg.Palette {
			char, r, g, b := c.pixelToChar(col)
			if c.Options.Mode == "ansi" {
				lut[i] = c.formatANSI(char, r, g, b)
			} else {
				lut[i] = string(char)
			}
		}

		for y := range targetHeight {
			srcY := y * heightIn / targetHeight
			for x := range c.Options.TargetWidth {
				srcX := x * widthIn / c.Options.TargetWidth
				builder.WriteString(lut[pImg.ColorIndexAt(srcX, srcY)])
			}
			builder.WriteByte('\n')
		}
		return builder.String()
	}

	for y := range targetHeight {
		srcY := y * heightIn / targetHeight
		for x := range c.Options.TargetWidth {
			srcX := x * widthIn / c.Options.TargetWidth
			char, r, g, b := c.pixelToChar(img.At(srcX, srcY))

			if c.Options.Mode == "ansi" {
				builder.WriteString(c.formatANSI(char, r, g, b))
			} else {
				builder.WriteByte(char)
			}
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (c *Converter) pixelToChar(p color.Color) (byte, uint8, uint8, uint8) {
	r, g, b, _ := p.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	lum := (uint32(r8)*19595 + uint32(g8)*38470 + uint32(b8)*7471) >> 16

	index := int(lum) * (len(charSet) - 1) / 255
	if index >= len(charSet) {
		index = len(charSet) - 1
	}

	return charSet[index], r8, g8, b8 
}

func Decode(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

func (c *Converter) formatANSI(char byte, r, g, b uint8) string {
	var bld strings.Builder
	bld.WriteString("\x1b[38;2;")
	bld.Write(strconv.AppendUint(nil, uint64(r), 10))
	bld.WriteByte(';')
	bld.Write(strconv.AppendUint(nil, uint64(g), 10))
	bld.WriteByte(';')
	bld.Write(strconv.AppendUint(nil, uint64(b), 10))
	bld.WriteByte('m')
	bld.WriteByte(char)
	bld.WriteString("\x1b[0m")
	return bld.String()
}

func (c *Converter) ConvertGIF(g *gif.GIF, frames chan<- string) {
	for _, frame := range g.Image {
		frames <- c.Convert(frame)
	}
	close(frames)
}
