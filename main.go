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

type Game struct {
	pole Pole
	x    int
	v    int
}

func (g *Game) Update() error {
	switch {
	case g.pole == PoleN && inpututil.IsKeyJustPressed(ebiten.KeyS):
		g.pole = PoleS
		g.v += 5
	case g.pole == PoleS && inpututil.IsKeyJustPressed(ebiten.KeyN):
		g.pole = PoleN
		g.v += 5
	default:
		g.v--
	}
	if g.v < 0 {
		g.v = 0
	}
	g.x += g.v
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	msg := fmt.Sprintf("Press S and N alternately!\nCurrent Pole: %s\nVelocity: %d\nPosition: %d", g.pole, g.v, g.x)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowTitle("Manual Linear Motor Car")
	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}
