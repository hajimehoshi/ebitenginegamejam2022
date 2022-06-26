// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

//go:embed bg.kage
var bgKage []byte

type gameSceneState int

const (
	gameSceneStateInit gameSceneState = iota
	gameSceneStateLogoFadeIn
	gameSceneStateLogoWait
	gameSceneStateBgFadeIn
	gameSceneStateTitleWait
	gameSceneStateGameCountDown
	gameSceneStateGamePlay
)

type GameScene struct {
	state      gameSceneState
	bgShader   *ebiten.Shader
	counter    int
	counterMax int
	gameState  GameState
}

func (g *GameScene) Update(sceneSwitcher SceneSwitcher) error {
	if g.bgShader == nil {
		s, err := ebiten.NewShader(bgKage)
		if err != nil {
			return err
		}
		g.bgShader = s
	}

	switch g.state {
	case gameSceneStateInit:
		g.state = gameSceneStateLogoFadeIn
		g.counterMax = ebiten.MaxTPS() / 2
		g.counter = g.counterMax
	case gameSceneStateLogoFadeIn:
		g.counter--
		if g.counter <= 0 {
			g.state = gameSceneStateLogoWait
			g.counterMax = ebiten.MaxTPS() / 2
			g.counter = g.counterMax
		}
	case gameSceneStateLogoWait:
		g.counter--
		if g.counter <= 0 {
			g.state = gameSceneStateBgFadeIn
			g.counterMax = ebiten.MaxTPS()
			g.counter = g.counterMax
			g.gameState.StartFixedVelocity()
		}
	case gameSceneStateBgFadeIn:
		g.counter--
		if g.counter <= 0 {
			g.state = gameSceneStateTitleWait
		}
	case gameSceneStateTitleWait:
		if g.counter > 0 {
			g.counter--
		}
		if !g.gameState.IsResetting() {
			if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
				g.gameState.Reset()
				g.counterMax = ebiten.MaxTPS() / 2
				g.counter = g.counterMax
			}
		}
		if g.gameState.CanStart() && g.counter <= 0 {
			g.state = gameSceneStateGameCountDown
			g.counterMax = ebiten.MaxTPS() * 3
			g.counter = g.counterMax
		}
	case gameSceneStateGameCountDown:
		g.counter--
		if g.counter <= 0 {
			g.state = gameSceneStateGamePlay
			g.gameState.Start()
		}
	}
	g.gameState.Update()
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	if g.state == gameSceneStateInit {
		return
	}

	// Render the background.
	switch g.state {
	case gameSceneStateBgFadeIn, gameSceneStateTitleWait, gameSceneStateGameCountDown, gameSceneStateGamePlay:
		sw, sh := screen.Size()
		alpha := float32(1)
		switch g.state {
		case gameSceneStateBgFadeIn:
			alpha = 1 - float32(g.counter)/float32(g.counterMax)
		}
		t := float32(g.gameState.PositionInMillimeter()) / 1000.0
		screen.DrawRectShader(sw, sh, g.bgShader, &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Time":  t,
				"Alpha": alpha,
			},
		})
	}

	// Render the title.
	switch g.state {
	case gameSceneStateLogoFadeIn, gameSceneStateLogoWait, gameSceneStateBgFadeIn, gameSceneStateTitleWait:
		sw, _ := screen.Size()
		alpha := 1.0
		switch g.state {
		case gameSceneStateLogoFadeIn:
			alpha = 1 - float64(g.counter)/float64(g.counterMax)
		case gameSceneStateTitleWait:
			if g.gameState.IsResetting() {
				alpha = float64(g.counter) / float64(g.counterMax)
			}
		}
		clr := color.RGBA{byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha)}
		for i, line := range []string{"Manual", "Linear", "Motor", "Car"} {
			f := spaceAgeBig
			r := text.BoundString(f, line)
			x := (sw-r.Dx())/2 - r.Min.X
			y := 144 + 144*i
			text.Draw(screen, line, f, x, y, clr)
		}
	case gameSceneStateGameCountDown:
		sw, _ := screen.Size()
		n := int(math.Ceil(float64(g.counter) / float64(ebiten.MaxTPS())))
		line := fmt.Sprintf("%d", n)
		f := spaceAgeBig
		r := text.BoundString(f, line)
		x := (sw-r.Dx())/2 - r.Min.X
		y := 144
		text.Draw(screen, line, f, x, y, color.White)
	}

	// Render the position and the velocity.
	switch g.state {
	case gameSceneStateTitleWait, gameSceneStateGameCountDown, gameSceneStateGamePlay:
		sw, sh := screen.Size()
		f := spaceAgeSmall
		r := text.BoundString(f, "km/h")
		offsetY := 32
		baseX := sw - (r.Dx() + r.Min.X)
		for i, line := range []string{"km/h", "m"} {
			x := baseX - 48
			y := sh + 72*i - 72 - offsetY
			text.Draw(screen, line, f, x, y, color.White)
		}

		v := g.gameState.VelocityInMeterPerHour()
		vstr := fmt.Sprintf("%d.%03d", v/1000, v%1000)
		p := g.gameState.PositionInMillimeter()
		pstr := fmt.Sprintf("%d.%03d", p/1000, p%1000)
		for j, line := range []string{vstr, pstr} {
			op := &ebiten.DrawImageOptions{}
			dotIndex := strings.Index(line, ".")
			for i, glyph := range text.AppendGlyphs(nil, f, line) {
				op.GeoM.Reset()
				const digitWidth = 108
				x := float64(baseX + (digitWidth-glyph.Image.Bounds().Dx())/2 - 72)
				switch {
				case i < dotIndex:
					x += float64(digitWidth*i + digitWidth*3/4 - digitWidth*len(line))
				case i == dotIndex:
					x += float64(digitWidth*i + digitWidth*3/8 - digitWidth*len(line))
				default:
					x += float64(digitWidth*i - digitWidth*len(line))
				}
				y := float64(sh+72*j-72-offsetY) + glyph.Y
				op.GeoM.Translate(x, y)
				screen.DrawImage(glyph.Image, op)
			}
		}
	}
}
