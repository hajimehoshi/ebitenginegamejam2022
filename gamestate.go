// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
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
	pole     Pole
	x        int // [mm]
	v        int // [m/h]
	vFixed   bool
	resetting bool
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func (g *GameState) Update() error {
	if !g.vFixed && !g.resetting {
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
	}
	if g.resetting {
		g.x -= max(5, g.x / ebiten.MaxTPS())
		if g.x < 0 {
			g.x = 0
		}
	} else {
		g.x += g.v * 1e3 / 3600 / ebiten.MaxTPS()
	}
	return nil
}

func (g *GameState) StartFixedVelocity() {
	g.x = 0
	g.v = 1000
	g.vFixed = true
	g.resetting = false
}

func (g *GameState) Reset() {
	g.vFixed = true
	g.resetting = true
}

func (g *GameState) IsResetting() bool {
	return g.resetting
}

func (g *GameState) CanStart() bool {
	return g.resetting == true && g.x == 0
}

func (g *GameState) VelocityInMeterPerHour() int {
	return g.v
}

func (g *GameState) PositionInMillimeter() int {
	return g.x
}
