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

type titleSceneState int

const (
	titleSceneStateInit titleSceneState = iota
	titleSceneStateLogoFadeIn
	titleSceneStateLogoWait
	titleSceneStateBgFadeIn
	titleSceneStateWait
)

type TitleScene struct {
	state      titleSceneState
	bgShader   *ebiten.Shader
	counter    int
	counterMax int
	time       int
}

func (t *TitleScene) Update(sceneSwitcher SceneSwitcher) error {
	if t.bgShader == nil {
		s, err := ebiten.NewShader(bgKage)
		if err != nil {
			return err
		}
		t.bgShader = s
	}

	switch t.state {
	case titleSceneStateInit:
		t.state = titleSceneStateLogoFadeIn
		t.counterMax = ebiten.MaxTPS() / 2
		t.counter = t.counterMax
	case titleSceneStateLogoFadeIn:
		t.counter--
		if t.counter <= 0 {
			t.state = titleSceneStateLogoWait
			t.counterMax = ebiten.MaxTPS() / 2
			t.counter = t.counterMax
		}
	case titleSceneStateLogoWait:
		t.counter--
		if t.counter <= 0 {
			t.state = titleSceneStateBgFadeIn
			t.counterMax = ebiten.MaxTPS()
			t.counter = t.counterMax
		}
	case titleSceneStateBgFadeIn:
		t.counter--
		if t.counter <= 0 {
			t.state = titleSceneStateWait
		}
	case titleSceneStateWait:
		if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
			sceneSwitcher.SwitchToGameScene()
			return nil
		}
	}
	t.time++
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	if t.state == titleSceneStateInit {
		return
	}

	switch t.state {
	case titleSceneStateBgFadeIn, titleSceneStateWait:
		sw, sh := screen.Size()
		alpha := float32(1)
		switch t.state {
		case titleSceneStateBgFadeIn:
			alpha = 1 - float32(t.counter)/float32(t.counterMax)
		}
		screen.DrawRectShader(sw, sh, t.bgShader, &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Time":  float32(t.time) / float32(ebiten.MaxTPS()),
				"Alpha": alpha,
			},
		})
	}

	sw, _ := screen.Size()
	alpha := 1.0
	switch t.state {
	case titleSceneStateLogoFadeIn:
		alpha = 1 - float64(t.counter)/float64(t.counterMax)
	}
	clr := color.RGBA{byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha), byte(0xff * alpha)}
	for i, line := range []string{"Manual", "Linear", "Motor", "Car"} {
		f := spaceAge
		r := text.BoundString(f, line)
		x := (sw-r.Dx())/2 - r.Min.X
		y := 144 + 144*i
		text.Draw(screen, line, f, x, y, clr)
	}
}
