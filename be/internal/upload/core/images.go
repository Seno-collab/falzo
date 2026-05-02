package core

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
)

func Decode(input []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(input))
	return img, format, err
}

func ToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}

	return dst
}

func Grayscale(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()

			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)
			aa := uint8(a >> 8)

			gray := uint8(0.299*float64(rr) + 0.587*float64(gg) + 0.114*float64(bb))

			dst.Set(x, y, color.RGBA{
				R: gray,
				G: gray,
				B: gray,
				A: aa,
			})
		}
	}

	return dst
}

func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer

	err := jpeg.Encode(&buf, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
