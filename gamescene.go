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

type GameScene struct {
	bgShader   *ebiten.Shader
	sequence   *Sequence
	gameState  GameState
	bgAlpha    float64
	logoAlpha  float64
	gaugeAlpha float64
	countDown  int
}

func (g *GameScene) Update(sceneSwitcher SceneSwitcher) error {
	if g.bgShader == nil {
		s, err := ebiten.NewShader(bgKage)
		if err != nil {
			return err
		}
		g.bgShader = s
	}

	if g.sequence == nil {
		g.sequence = &Sequence{}
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			g.bgAlpha = 0
			g.logoAlpha = float64(counter) / float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()/2))
		g.sequence.AddTask(func () error {
			g.gameState.StartFixedVelocity()
			return TaskEnded
		})
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			g.bgAlpha = float64(counter) / float64(maxCounter)
			g.logoAlpha = 1
			g.gaugeAlpha = float64(counter) / float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()))
		g.sequence.AddTask(func() error {
			if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
				g.gameState.Reset()
				return TaskEnded
			}
			return nil
		})
		g.sequence.AddTask(NewAllTask(func() error {
			// Wait the game state is ready to start.
			if !g.gameState.CanStart() {
				return nil
			}
			return TaskEnded
		}, NewTimerTask(func(counter int, maxCounter int) error {
			// Fade out the logo.
			g.bgAlpha = 1
			g.logoAlpha = 1 - float64(counter)/float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()/2)))
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			g.countDown = int(math.Ceil(float64(maxCounter-counter) / float64(ebiten.MaxTPS())))
			return nil
		}, ebiten.MaxTPS()*3))
		g.sequence.AddTask(func() error {
			g.gameState.Start()
			return TaskEnded
		})
	}

	g.sequence.Update()
	g.gameState.Update()
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	if g.sequence == nil {
		return
	}

	// Render the background.
	if g.bgAlpha > 0 {
		sw, sh := screen.Size()
		t := float32(g.gameState.PositionInMillimeter()) / 1000.0
		v := float32(g.gameState.VelocityInMeterPerHour()) / 1000.0
		screen.DrawRectShader(sw, sh, g.bgShader, &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Pos":      t,
				"Velocity": v,
				"Alpha":    float32(g.bgAlpha),
			},
		})
	}

	// Render the title.
	if g.logoAlpha > 0 {
		sw, _ := screen.Size()
		alpha := g.logoAlpha
		clr := color.RGBA{byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha)}
		for i, line := range []string{"Manual", "Linear", "Motor", "Car"} {
			f := spaceAgeBig
			r := text.BoundString(f, line)
			x := (sw-r.Dx())/2 - r.Min.X
			y := 144 + 144*i
			text.Draw(screen, line, f, x, y, clr)
		}
	}
	if g.countDown > 0 {
		sw, _ := screen.Size()
		line := fmt.Sprintf("%d", g.countDown)
		f := spaceAgeBig
		r := text.BoundString(f, line)
		x := (sw-r.Dx())/2 - r.Min.X
		y := 144
		text.Draw(screen, line, f, x, y, color.White)
	}

	// Render the position and the velocity.
	if g.gaugeAlpha > 0 {
		sw, sh := screen.Size()
		f := spaceAgeSmall
		r := text.BoundString(f, "km/h")
		offsetY := 32
		baseX := sw - (r.Dx() + r.Min.X)
		alpha := g.gaugeAlpha
		clr := color.RGBA{byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha)}
		for i, line := range []string{"km/h", "m"} {
			x := baseX - 48
			y := sh + 72*i - 72 - offsetY
			text.Draw(screen, line, f, x, y, clr)
		}

		v := g.gameState.VelocityInMeterPerHour()
		vstr := fmt.Sprintf("%d.%03d", v/1000, v%1000)
		p := g.gameState.PositionInMillimeter()
		pstr := fmt.Sprintf("%d.%03d", p/1000, p%1000)
		for j, line := range []string{vstr, pstr} {
			op := &ebiten.DrawImageOptions{}
			dotIndex := strings.Index(line, ".")
			for i, glyph := range text.AppendGlyphs(nil, f, line) {
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
				op.GeoM.Reset()
				op.GeoM.Translate(x, y)
				op.ColorM.Reset()
				op.ColorM.Scale(1, 1, 1, alpha)
				screen.DrawImage(glyph.Image, op)
			}
		}
	}

	// Render the time.
}
