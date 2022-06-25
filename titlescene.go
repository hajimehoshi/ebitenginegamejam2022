// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type TitleScene struct {
}

func (t *TitleScene) Update(sceneSwitcher SceneSwitcher) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		sceneSwitcher.SwitchToGameScene()
		return nil
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
}
