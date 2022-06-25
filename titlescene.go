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

//go:embed titlebg.kage
var titlebgKage []byte

type TitleScene struct {
	counter  int
	bgShader *ebiten.Shader
}

func (t *TitleScene) Update(sceneSwitcher SceneSwitcher) error {
	if t.bgShader == nil {
		s, err := ebiten.NewShader(titlebgKage)
		if err != nil {
			return err
		}
		t.bgShader = s
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		sceneSwitcher.SwitchToGameScene()
		return nil
	}

	t.counter++
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	if t.bgShader == nil {
		return
	}

	sw, sh := screen.Size()
	screen.DrawRectShader(sw, sh, t.bgShader, &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"Time": float32(t.counter) / float32(ebiten.MaxTPS()),
		},
	})

	for i, line := range []string{"Manual", "Linear", "Motor", "Car"} {
		f := robotoBold
		r := text.BoundString(f, line)
		x := (sw - r.Dx()) / 2 - r.Min.X
		y := 144 + 144 * i
		text.Draw(screen, line, robotoBold, x, y, color.White)
	}
}
