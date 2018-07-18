package main

import (
	"fmt"
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
	title, w, h, images, game := runLoader(program)
	runGame(title, w, h, images, game)
}

func runLoader(loader *runtime.Box) (title string, w, h int, images []*ebiten.Image, game *runtime.Box) {
	for {
		switch loader.Alternative() {
		case 0: // start
			title, w, h, game := loader.Field(0), loader.Field(1), loader.Field(2), loader.Field(3)
			return title.String(), int(w.Int().Int64()), int(h.Int().Int64()), images, game
		case 1: // load-image
			path := loader.Field(0).String()
			ebitenImg, _, err := ebitenutil.NewImageFromFile(path, ebiten.FilterLinear)
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
		}
	}
}

func runGame(title string, w, h int, images []*ebiten.Image, game *runtime.Box) {
	update := func(screen *ebiten.Image) error {
		screen.Fill(color.Black)
		switch game.Alternative() {
		case 0: // done
			return io.EOF
		case 1: // frame
			sprites := game.Field(0).List()
			for _, sprite := range sprites {
				var (
					img              = sprite.Field(0)
					imgId            = img.Field(0).Int().Int64()
					pos              = sprite.Field(1)
					posX, posY       = pos.Field(0).Float(), pos.Field(1).Float()
					anchor           = sprite.Field(2)
					anchorX, anchorY = anchor.Field(0).Float(), anchor.Field(1).Float()
					angle            = sprite.Field(3).Float()
					scale            = sprite.Field(4).Float()
				)
				var opts ebiten.DrawImageOptions
				opts.GeoM.Translate(-anchorX, -anchorY)
				opts.GeoM.Rotate(angle)
				opts.GeoM.Scale(scale, scale)
				opts.GeoM.Translate(posX, posY)
				screen.DrawImage(images[imgId], &opts)
			}
			input := runtime.MkRecord(
				runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyLeft)),
				runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyRight)),
				runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyUp)),
				runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeyDown)),
				runtime.MkBool(ebiten.IsKeyPressed(ebiten.KeySpace)),
			)
			game = game.Field(1).Apply(input)
			return nil
		}
		panic("unreachable")
	}
	ebiten.Run(update, w, h, 1, title)
}
