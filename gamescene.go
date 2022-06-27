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
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

//go:embed bg.kage
var bgKage []byte

type GameScene struct {
	bgShader       *ebiten.Shader
	sequence       *Sequence
	gameState      GameState
	bgAlpha        float64
	logoAlpha      float64
	showPressSpace bool
	gaugeAlpha     float64
	showRecord     bool
	countDown      int
	topVelocity    int
	lastPosition   int

	audioContext  *audio.Context
	bgmPlayer     *audio.Player
	seStartPlayer *audio.Player
	seEndPlayer   *audio.Player
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
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			if counter == 0 {
				g.gameState.StartDemo()
			}
			g.bgAlpha = float64(counter) / float64(maxCounter)
			g.logoAlpha = 1
			g.gaugeAlpha = float64(counter) / float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()))
	}
	var addGameLoopTasks func()
	addGameLoopTasks = func() {
		g.sequence.AddTask(func() error {
			g.showPressSpace = true
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.showPressSpace = false
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
			g.logoAlpha = 0
			return nil
		}, ebiten.MaxTPS()*3))
		g.sequence.AddTask(func() error {
			g.countDown = 0
			g.gameState.Start()
			if err := g.seStartPlayer.Rewind(); err != nil {
				return err
			}
			g.seStartPlayer.Play()
			return TaskEnded
		})
		g.sequence.AddTask(func() error {
			if g.gameState.IsPlaying() {
				return nil
			}
			g.showRecord = true
			g.topVelocity, g.lastPosition = g.gameState.Record()
			g.gameState.StartDemo()
			if err := g.seEndPlayer.Rewind(); err != nil {
				return err
			}
			g.seEndPlayer.Play()
			return TaskEnded
		})
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			// Cool time
			return nil
		}, ebiten.MaxTPS()))
		g.sequence.AddTask(func() error {
			g.showPressSpace = true
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.showRecord = false
				return TaskEnded
			}
			return nil
		})
		g.sequence.AddTask(NewTimerTask(func(counter int, maxCounter int) error {
			g.logoAlpha = float64(counter) / float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()/2))
		g.sequence.AddTask(func() error {
			addGameLoopTasks()
			return TaskEnded
		})
	}

	if g.audioContext == nil {
		const sampleRate = 48000
		g.audioContext = audio.NewContext(sampleRate)
		{
			f, err := resourceFS.Open("bgm.ogg")
			if err != nil {
				return err
			}
			defer f.Close()

			decoded, err := vorbis.DecodeWithSampleRate(sampleRate, f)
			if err != nil {
				return err
			}

			loop := audio.NewInfiniteLoop(decoded, decoded.Length())
			p, err := g.audioContext.NewPlayer(loop)
			if err != nil {
				return err
			}
			g.bgmPlayer = p
			g.bgmPlayer.SetVolume(0.8)
			g.bgmPlayer.Play()
		}
		{
			f, err := resourceFS.Open("start.wav")
			if err != nil {
				return err
			}
			defer f.Close()

			decoded, err := wav.DecodeWithSampleRate(sampleRate, f)
			if err != nil {
				return err
			}

			p, err := g.audioContext.NewPlayer(decoded)
			g.seStartPlayer = p
			g.seStartPlayer.SetVolume(0.8)
		}
		{
			f, err := resourceFS.Open("end.wav")
			if err != nil {
				return err
			}
			defer f.Close()

			decoded, err := wav.DecodeWithSampleRate(sampleRate, f)
			if err != nil {
				return err
			}

			p, err := g.audioContext.NewPlayer(decoded)
			g.seEndPlayer = p
			g.seEndPlayer.SetVolume(0.8)
		}
	}

	addGameLoopTasks()

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
		t := float32(g.gameState.PositionInMillimeter() % 1000) / 1000.0
		v := float32(g.gameState.VelocityInMeterPerHour()) / 1000.0
		screen.DrawRectShader(sw, sh, g.bgShader, &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Pos":      t, // [0, 1]
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
			x := float64(baseX - 72)
			y := float64(sh + 72*j - 72 - offsetY)
			renderNumberWithDecimalPoint(screen, line, x, y, alpha)
		}
	}

	// Render the time.
	if c := g.gameState.Counter(); c > 0 {
		str := fmt.Sprintf("%.2f", float64(c)/float64(ebiten.MaxTPS()))
		x := 480.0
		y := 96.0
		renderNumberWithDecimalPoint(screen, str, x, y, 1)
	}

	// Render the record
	if g.showRecord {
		sw, _ := screen.Size()
		v := g.topVelocity
		p := g.lastPosition
		lines := []string{
			"Record",
			"Top Speed",
			fmt.Sprintf("%d.%03d km/h", v/1000, v%1000),
			"Distance",
			fmt.Sprintf("%d.%03d m", p/1000, p%1000),
		}
		for i, line := range lines {
			f := spaceAgeSmall
			r := text.BoundString(f, line)
			x := (sw-r.Dx())/2 - r.Min.X
			y := 144 + 96*i
			text.Draw(screen, line, f, x, y, color.White)
		}
	}

	if g.showPressSpace {
		sw, sh := screen.Size()
		lines := []string{
			"Press Space Key",
		}
		for _, line := range lines {
			f := spaceAgeSmall
			r := text.BoundString(f, line)
			x := (sw-r.Dx())/2 - r.Min.X
			y := sh - 144 - 96
			text.Draw(screen, line, f, x, y, color.RGBA{0xa0, 0xa0, 0xa0, 0xff})
		}
	} else if g.gameState.ShouldShowGuide() {
		sw, sh := screen.Size()
		var str string
		switch g.gameState.Pole() {
		case PoleS:
			str = "Press N Key"
		case PoleN:
			str = "Press S Key"
		}
		lines := []string{
			str,
		}
		for _, line := range lines {
			f := spaceAgeSmall
			r := text.BoundString(f, line)
			x := (sw-r.Dx())/2 - r.Min.X
			y := sh - 144 - 96
			text.Draw(screen, line, f, x, y, color.RGBA{0xa0, 0xa0, 0xa0, 0xff})
		}
	}
}

func renderNumberWithDecimalPoint(dst *ebiten.Image, str string, ox, oy float64, alpha float64) {
	// TODO: Define a new font.Face to overwrite the kerning.
	f := spaceAgeSmall
	op := &ebiten.DrawImageOptions{}
	dotIndex := strings.Index(str, ".")
	for i, glyph := range text.AppendGlyphs(nil, f, str) {
		const digitWidth = 108
		x := ox
		switch {
		case i < dotIndex:
			x += float64(digitWidth*i + digitWidth*3/4 - digitWidth*len(str))
		case i == dotIndex:
			x += float64(digitWidth*i + digitWidth*3/8 - digitWidth*len(str))
		default:
			x += float64(digitWidth*i - digitWidth*len(str))
		}
		x += float64(digitWidth-glyph.Image.Bounds().Dx()) / 2
		y := oy + glyph.Y
		op.GeoM.Reset()
		op.GeoM.Translate(x, y)
		op.ColorM.Reset()
		op.ColorM.Scale(1, 1, 1, alpha)
		dst.DrawImage(glyph.Image, op)
	}
}
