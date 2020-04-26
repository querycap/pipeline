package handler

import (
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/storage"
)

func Handler(t operator.Transfer) error {
	for t.Scan() {
		var img image.Image

		err := pipeline.ReadNext(t, func(r io.Reader) error {
			i, _, err := image.Decode(r)
			if err != nil {
				return err
			}
			img = i
			return nil
		})

		if err != nil {
			return err
		}

		bounds := img.Bounds()
		dx := bounds.Dx()
		dy := bounds.Dy()

		nextImg := image.NewRGBA(bounds)

		for i := 0; i < dx; i++ {
			for j := 0; j < dy; j++ {
				_, g, _, a := img.At(i, j).RGBA()
				nextG := uint8(g >> 8)
				nextA := uint8(a >> 8)

				nextImg.SetRGBA(i, j, color.RGBA{nextG, nextG, nextG, nextA})
			}
		}

		err = t.Put(storage.WithContentType("image/png")(storage.WriteTo(func(w io.Writer) (int64, error) {
			return -1, png.Encode(w, nextImg)
		})))

		if err != nil {
			return err
		}
	}

	return nil
}
