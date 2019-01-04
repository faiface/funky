package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math/big"
	"os"

	"github.com/faiface/funky"
	"github.com/faiface/funky/runtime"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

func main() {
	program, cleanup := funky.Run("main")
	defer cleanup()
	title, w, h, images, loop := runLoader(program)
	runLoop(title, w, h, images, loop)
}

func runLoader(loader *runtime.Value) (title string, w, h int, images []*ebiten.Image, loop *runtime.Value) {
	for {
		switch loader.Alternative() {
		case 0: // start
			title, w, h, loop := loader.Field(0), loader.Field(1), loader.Field(2), loader.Field(3)
			return title.String(), int(w.Int().Int64()), int(h.Int().Int64()), images, loop
		case 1: // load-image
			path := loader.Field(0).String()
			ebitenImg, _, err := ebitenutil.NewImageFromFile(path, ebiten.FilterNearest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			id := len(images)
			width, height := ebitenImg.Size()
			img := runtime.MkRecord(
				runtime.MkInt(big.NewInt(int64(id))),
				runtime.MkInt(big.NewInt(int64(width))),
				runtime.MkInt(big.NewInt(int64(height))),
			)
			images = append(images, ebitenImg)
			loader = loader.Field(1).Apply(img)
			continue
		case 2: // make-image
			width := int(loader.Field(0).Int().Int64())
			height := int(loader.Field(1).Int().Int64())
			fn := loader.Field(2)
			rgba := image.NewRGBA(image.Rect(0, 0, width, height))
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					xyColor := fn.Apply(runtime.MkRecord(
						runtime.MkFloat(float64(x)),
						runtime.MkFloat(float64(y)),
					))
					r := clamp(0, 1, xyColor.Field(0).Float())
					g := clamp(0, 1, xyColor.Field(1).Float())
					b := clamp(0, 1, xyColor.Field(2).Float())
					a := clamp(0, 1, xyColor.Field(3).Float())
					rgba.SetRGBA(x, y, color.RGBA{
						R: uint8(r * 255),
						G: uint8(g * 255),
						B: uint8(b * 255),
						A: uint8(a * 255),
					})
				}
			}
			ebitenImg, err := ebiten.NewImageFromImage(rgba, ebiten.FilterNearest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			id := len(images)
			img := runtime.MkRecord(
				runtime.MkInt(big.NewInt(int64(id))),
				runtime.MkInt(big.NewInt(int64(width))),
				runtime.MkInt(big.NewInt(int64(height))),
			)
			images = append(images, ebitenImg)
			loader = loader.Field(3).Apply(img)
		}
	}
}

func runLoop(title string, w, h int, images []*ebiten.Image, loop *runtime.Value) {
	update := func(screen *ebiten.Image) error {
		screen.Fill(color.Black)
		for {
			switch loop.Alternative() {
			case 0: // quit
				return io.EOF
			case 1: // fill
				bgColor := loop.Field(0)
				r := clamp(0, 1, bgColor.Field(0).Float())
				g := clamp(0, 1, bgColor.Field(1).Float())
				b := clamp(0, 1, bgColor.Field(2).Float())
				a := clamp(0, 1, bgColor.Field(3).Float())
				screen.Fill(color.RGBA{
					R: uint8(r * 255),
					G: uint8(g * 255),
					B: uint8(b * 255),
					A: uint8(a * 255),
				})
				loop = loop.Field(1)
			case 2: // draw
				var (
					sprite           = loop.Field(0)
					img              = sprite.Field(0)
					imgId            = img.Field(0).Int().Int64()
					pos              = sprite.Field(1)
					posX, posY       = pos.Field(0).Float(), pos.Field(1).Float()
					anchor           = sprite.Field(2)
					anchorX, anchorY = anchor.Field(0).Float(), anchor.Field(1).Float()
					angle            = sprite.Field(3).Float()
					scale            = sprite.Field(4).Float()
					mask             = sprite.Field(5)
					maskR            = mask.Field(0).Float()
					maskG            = mask.Field(1).Float()
					maskB            = mask.Field(2).Float()
					maskA            = mask.Field(3).Float()
				)
				var opts ebiten.DrawImageOptions
				opts.GeoM.Translate(-anchorX, -anchorY)
				opts.GeoM.Rotate(angle)
				opts.GeoM.Scale(scale, scale)
				opts.GeoM.Translate(posX, posY)
				opts.ColorM.Scale(maskR, maskG, maskB, maskA)
				screen.DrawImage(images[imgId], &opts)
				loop = loop.Field(1)
			case 3: // present
				input := runtime.MkRecord(
					runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyLeft)),
					runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyRight)),
					runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyUp)),
					runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyDown)),
					runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeySpace)),
				)
				loop = loop.Field(0).Apply(input)
				return nil
			default:
				panic("unreachable")
			}
		}
	}
	ebiten.Run(update, w, h, 1, title)
}

func clamp(min, max, x float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
