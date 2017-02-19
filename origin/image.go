package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand"
)

// TODO: say which image we are
// http://stackoverflow.com/questions/38299930/how-to-add-a-simple-text-label-to-an-image-in-go

func genImage(id int) []byte {
	rander := rand.New(rand.NewSource(int64(id)))
	bounds := image.Rect(0, 0, 100, 50)
	img := image.NewRGBA(bounds)
	for i := 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			img.Set(i, j, color.RGBA{
				uint8(rander.Intn(255)),
				uint8(rander.Intn(255)),
				uint8(rander.Intn(255)),
				255,
			})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
