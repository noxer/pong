package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	ttf "github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/noxer/pong/res"
	"golang.org/x/image/font"
)

var (
	red       = color.RGBA{255, 0, 0, 255}
	green     = color.RGBA{0, 255, 0, 255}
	white     = color.RGBA{255, 255, 255, 255}
	black     = color.RGBA{0, 0, 0, 255}
	gray      = color.Gray{100}
	lightGray = color.Gray{200}
)

type gameState struct {
	Round     int
	Ball      *ebiten.Image
	Rotation  float64
	Position  point
	Direction point
	Area      area
	Players   [2]player
	Fonts     fonts
}

type player struct {
	Name     string
	Avatar   *ebiten.Image
	Points   int
	Racket   *ebiten.Image
	Position point
	Height   float64
}

type point struct {
	X, Y float64
}

type fonts struct {
	ArcadeN        font.Face
	MPlus1pRegular font.Face
}

func (p point) Add(p2 point) point {
	return point{
		X: p.X + p2.X,
		Y: p.Y + p2.Y,
	}
}

func (p point) In(a area) bool {
	return p.X >= a.TopLeft.X && p.X <= a.BottomRight.X && p.Y >= a.TopLeft.Y && p.Y <= a.BottomRight.Y
}

func (p point) RightOf(p2 point) bool {
	return p.X > p2.X
}

func (p point) LeftOf(p2 point) bool {
	return p.X < p2.X
}

func (p point) Above(p2 point) bool {
	return p.Y < p2.Y
}

func (p point) Below(p2 point) bool {
	return p.Y > p2.Y
}

type area struct {
	TopLeft     point
	BottomRight point
}

func (a area) Center() point {
	return point{
		X: a.CenterX(),
		Y: a.CenterY(),
	}
}

func (a area) CenterX() float64 {
	return a.TopLeft.X + (a.BottomRight.X-a.TopLeft.X)/2
}

func (a area) CenterY() float64 {
	return a.TopLeft.Y + (a.BottomRight.Y-a.TopLeft.Y)/2
}

func (a area) Draw(dst *ebiten.Image, c color.Color) {
	ebitenutil.DrawRect(dst, a.TopLeft.X, a.TopLeft.Y, a.BottomRight.X-a.TopLeft.X, a.BottomRight.Y-a.TopLeft.Y, c)
}

var state gameState

// update is called every frame (1/60 [s]).
func update(screen *ebiten.Image) error {
	// Write your game's logical update.
	state.Rotation += .05
	if state.Rotation > 360.0 {
		state.Rotation -= 360.0
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		state.Players[0].Position.Y -= 1.0
		if state.Players[0].Position.Y < state.Area.TopLeft.Y {
			state.Players[0].Position.Y = state.Area.TopLeft.Y
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		state.Players[0].Position.Y += 1.0
		if state.Players[0].Position.Y > state.Area.BottomRight.Y-state.Players[0].Height {
			state.Players[0].Position.Y = state.Area.BottomRight.Y - state.Players[0].Height
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		state.Players[1].Position.Y -= 1.0
		if state.Players[1].Position.Y < state.Area.TopLeft.Y {
			state.Players[1].Position.Y = state.Area.TopLeft.Y
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		state.Players[1].Position.Y += 1.0
		if state.Players[1].Position.Y > state.Area.BottomRight.Y-state.Players[1].Height {
			state.Players[1].Position.Y = state.Area.BottomRight.Y - state.Players[1].Height
		}
	}

	if state.Direction.X > 0 {
		if state.Position.X < state.Area.BottomRight.X-25 && state.Position.X+state.Direction.X >= state.Area.BottomRight.X-25 {
			if state.Position.Y >= state.Players[1].Position.Y && state.Position.Y < state.Players[1].Position.Y+state.Players[1].Height {
				state.Direction.X = -state.Direction.X
				p := ((state.Position.Y - state.Players[1].Position.Y) / state.Players[1].Height) - 0.5
				state.Direction.Y += p
			}
		}
	} else {
		if state.Position.X > state.Area.TopLeft.X+25 && state.Position.X+state.Direction.X <= state.Area.TopLeft.X+25 {
			if state.Position.Y >= state.Players[0].Position.Y && state.Position.Y < state.Players[0].Position.Y+state.Players[0].Height {
				state.Direction.X = -state.Direction.X
				p := ((state.Position.Y - state.Players[0].Position.Y) / state.Players[0].Height) - 0.5
				state.Direction.Y += p
			}
		}
	}

	state.Position = state.Position.Add(state.Direction)
	if state.Position.Y <= state.Area.TopLeft.Y+5 {
		state.Direction.Y = -state.Direction.Y
		state.Position.Y += (state.Area.TopLeft.Y + 5) - state.Position.Y
	}
	if state.Position.Y >= state.Area.BottomRight.Y-5 {
		state.Direction.Y = -state.Direction.Y
		state.Position.Y += (state.Area.BottomRight.Y - 5) - state.Position.Y
	}

	if state.Position.X <= state.Area.TopLeft.X+5 {
		state.Players[1].Points++
		state.Round++
		state.Position = state.Area.Center()
		state.Direction.X = 1
		state.Direction.Y = rand.Float64() - 0.5
	}
	if state.Position.X >= state.Area.BottomRight.X-5 {
		state.Players[0].Points++
		state.Round++
		state.Position = state.Area.Center()
		state.Direction.X = -1
		state.Direction.Y = rand.Float64() - 0.5
	}

	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Write your game's rendering.
	screen.Fill(red)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(10, 10)
	screen.DrawImage(state.Players[0].Avatar, op)

	op.GeoM.Reset()
	op.GeoM.Translate(310-32, 10)
	screen.DrawImage(state.Players[1].Avatar, op)

	state.Area.Draw(screen, gray)

	op.GeoM.Reset()
	op.GeoM.Translate(-5, -5)
	op.GeoM.Rotate(state.Rotation)
	op.GeoM.Translate(state.Position.X, state.Position.Y)
	screen.DrawImage(state.Ball, op)

	w, h := textSize(state.Fonts.ArcadeN, state.Players[1].Name)
	drawText(screen, state.Fonts.ArcadeN, state.Players[0].Name, 45, 10+h, white)
	w, h = textSize(state.Fonts.ArcadeN, state.Players[1].Name)
	drawText(screen, state.Fonts.ArcadeN, state.Players[1].Name, 320-45-w, 10+h, black)

	score1 := strconv.Itoa(state.Players[0].Points)
	score2 := strconv.Itoa(state.Players[1].Points)

	drawText(screen, state.Fonts.MPlus1pRegular, score1, 45, 42, white)
	w, _ = textSize(state.Fonts.MPlus1pRegular, score2)
	drawText(screen, state.Fonts.MPlus1pRegular, score2, 320-45-w, 42, black)

	round := strconv.Itoa(state.Round)
	drawTextCenter(screen, state.Fonts.MPlus1pRegular, round, 320/2, 26, green)

	op.GeoM.Reset()
	op.GeoM.Translate(state.Players[0].Position.X-10, state.Players[0].Position.Y)
	screen.DrawImage(state.Players[0].Racket, op)

	op.GeoM.Reset()
	op.GeoM.Translate(state.Players[1].Position.X, state.Players[1].Position.Y)
	screen.DrawImage(state.Players[1].Racket, op)

	return nil
}

func drawText(dst *ebiten.Image, f font.Face, t string, x, y int, c color.Color) {
	text.Draw(dst, t, f, x, y, c)
}

func textSize(f font.Face, t string) (int, int) {
	bounds, _ := font.BoundString(f, t)
	return (bounds.Max.X - bounds.Min.X).Ceil(), (bounds.Max.Y - bounds.Min.Y).Ceil()
}

func drawTextCenter(dst *ebiten.Image, f font.Face, t string, x, y int, c color.Color) {
	w, h := textSize(f, t)
	drawText(dst, f, t, x-w/2, y+h/2, c)
}

func initGame() error {
	png, _, err := image.Decode(bytes.NewReader(res.Ball))
	if err != nil {
		return err
	}
	state.Round = 1
	img, err := ebiten.NewImageFromImage(png, ebiten.FilterDefault)
	if err != nil {
		return err
	}
	state.Ball = img
	state.Area = area{
		TopLeft:     point{X: 10, Y: 50},
		BottomRight: point{X: 310, Y: 230},
	}
	state.Position = state.Area.Center()
	state.Direction = point{
		X: 1,
		Y: rand.Float64(),
	}

	for i := range state.Players {
		if err = initPlayer(i); err != nil {
			return err
		}
	}

	if err = initFonts(); err != nil {
		return err
	}

	return nil
}

func initPlayer(n int) error {
	ava, err := ebiten.NewImage(32, 32, ebiten.FilterDefault)
	if err != nil {
		return err
	}

	rac, err := ebiten.NewImage(10, 30, ebiten.FilterDefault)
	if err != nil {
		return err
	}

	state.Players[n] = player{
		Name:     fmt.Sprintf("Player %d", n+1),
		Avatar:   ava,
		Points:   0,
		Racket:   rac,
		Position: point{Y: state.Area.CenterY()},
		Height:   30,
	}

	if n == 0 {
		state.Players[0].Avatar.Fill(white)
		state.Players[0].Racket.Fill(white)
		state.Players[0].Position.X = state.Area.TopLeft.X + 20
	} else {
		state.Players[1].Avatar.Fill(black)
		state.Players[1].Racket.Fill(black)
		state.Players[1].Position.X = state.Area.BottomRight.X - 20
	}

	return nil
}

func initFonts() error {
	f, err := truetype.Parse(ttf.ArcadeN_ttf)
	if err != nil {
		return err
	}

	state.Fonts.ArcadeN = truetype.NewFace(f, &truetype.Options{
		Size:    10,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	f, err = truetype.Parse(ttf.MPlus1pRegular_ttf)
	if err != nil {
		return err
	}

	state.Fonts.MPlus1pRegular = truetype.NewFace(f, &truetype.Options{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return nil
}

func main() {
	err := initGame()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	// Call ebiten.Run to start your game loop.
	if err := ebiten.Run(update, 320, 240, 2, "Pong"); err != nil {
		log.Fatal(err)
	}
}
