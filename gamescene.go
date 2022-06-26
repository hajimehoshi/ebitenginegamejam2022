// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	_ "embed"
	"image/color"

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
	gameSceneStateWait
)

type GameScene struct {
	state      gameSceneState
	bgShader   *ebiten.Shader
	counter    int
	counterMax int
	time       int
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
		}
	case gameSceneStateBgFadeIn:
		g.counter--
		if g.counter <= 0 {
			g.state = gameSceneStateWait
		}
	case gameSceneStateWait:
		if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
			return nil
		}
	}
	g.time++
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	if g.state == gameSceneStateInit {
		return
	}

	switch g.state {
	case gameSceneStateBgFadeIn, gameSceneStateWait:
		sw, sh := screen.Size()
		alpha := float32(1)
		switch g.state {
		case gameSceneStateBgFadeIn:
			alpha = 1 - float32(g.counter)/float32(g.counterMax)
		}
		screen.DrawRectShader(sw, sh, g.bgShader, &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Time":  float32(g.time) / float32(ebiten.MaxTPS()),
				"Alpha": alpha,
			},
		})
	}

	sw, _ := screen.Size()
	alpha := 1.0
	switch g.state {
	case gameSceneStateLogoFadeIn:
		alpha = 1 - float64(g.counter)/float64(g.counterMax)
	}
	clr := color.RGBA{byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha)}
	for i, line := range []string{"Manual", "Linear", "Motor", "Car"} {
		f := spaceAgeBig
		r := text.BoundString(f, line)
		x := (sw-r.Dx())/2 - r.Min.X
		y := 144 + 144*i
		text.Draw(screen, line, f, x, y, clr)
	}
}
