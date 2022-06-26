// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Pole int

const (
	PoleN Pole = iota
	PoleS
)

func (p Pole) String() string {
	switch p {
	case PoleN:
		return "N"
	case PoleS:
		return "S"
	default:
		panic("invalid pole")
	}
}

type GameState struct {
	pole Pole
	x    int // [mm]
	v    int // [m/h]
}

func (g *GameState) Update(sceneSwitcher SceneSwitcher) error {
	switch {
	case g.pole == PoleN && inpututil.IsKeyJustPressed(ebiten.KeyS):
		g.pole = PoleS
		g.v += 25000
	case g.pole == PoleS && inpututil.IsKeyJustPressed(ebiten.KeyN):
		g.pole = PoleN
		g.v += 25000
	default:
		g.v -= 2500
	}
	if g.v < 0 {
		g.v = 0
	}
	g.x += g.v * 1e3 / 3600 / ebiten.MaxTPS()
	return nil
}

func (g *GameState) Draw(screen *ebiten.Image) {
	msg := fmt.Sprintf("Press S and N alternately!\nCurrent Pole: %s\nVelocity: %d.%03d [km/h]\nPosition: %d.%03d [m]", g.pole, g.v/1000, g.v%1000, g.x/1000, g.x%1000)
	ebitenutil.DebugPrint(screen, msg)
}
