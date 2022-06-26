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

type GameStateMode int

const (
	GameStateModeWait GameStateMode = iota
	GameStateModeDemo
	GameStateModeResetting
	GameStateModePlay
)

type GameState struct {
	pole    Pole
	x       int // [mm]
	v       int // [m/h]
	mode    GameStateMode
	counter int
	lastX   int
	topV    int
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func (g *GameState) Update() error {
	switch g.mode {
	case GameStateModePlay:
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
	case GameStateModeResetting:
		g.v = 0
		g.x -= max(10, g.x/(ebiten.MaxTPS()/2))
		if g.x < 0 {
			g.x = 0
		}
	case GameStateModeDemo:
		if g.v > 1000 {
			g.v -= 2500
		}
		if g.v < 1000 {
			g.v = 1000
		}
	default:
		g.v -= 2500
		if g.v < 0 {
			g.v = 0
		}
	}
	if g.mode != GameStateModeResetting {
		g.x += g.v * 1e3 / (3600 * ebiten.MaxTPS())
	}
	if g.counter > 0 {
		g.counter--
		if g.topV < g.v {
			g.topV = g.v
		}
		if g.counter == 0 && g.mode == GameStateModePlay {
			g.lastX = g.x
			g.mode = GameStateModeWait
		}
	}
	return nil
}

func (g *GameState) StartDemo() {
	g.mode = GameStateModeDemo
}

func (g *GameState) Reset() {
	g.mode = GameStateModeResetting
}

func (g *GameState) CanStart() bool {
	return g.mode == GameStateModeResetting && g.x == 0
}

func (g *GameState) Start() {
	if !g.CanStart() {
		return
	}
	g.pole = PoleS
	g.mode = GameStateModePlay
	g.counter = ebiten.MaxTPS() * 20
	g.lastX = 0
	g.topV = 0
}

func (g *GameState) IsPlaying() bool {
	return g.mode == GameStateModePlay
}

func (g *GameState) Record() (int, int) {
	return g.topV, g.lastX
}

func (g *GameState) Counter() int {
	return g.counter
}

func (g *GameState) VelocityInMeterPerHour() int {
	return g.v
}

func (g *GameState) PositionInMillimeter() int {
	return g.x
}
