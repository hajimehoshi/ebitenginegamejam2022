// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type splashSceneState int

const (
	splashSceneStateInit splashSceneState = iota
	splashSceneStateFadeIn
	splashSceneStateWait
	splashSceneStateFadeOut
	splashSceneStateQuit
)

type SplashScene struct {
	state      splashSceneState
	splashImg  *ebiten.Image
	counter    int
	counterMax int
}

func (s *SplashScene) Update(sceneSwitcher SceneSwitcher) error {
	if s.splashImg == nil {
		f, err := resourceFS.Open("splash_1920x1080_black.png")
		if err != nil {
			return err
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			return err
		}
		s.splashImg = ebiten.NewImageFromImage(img)
	}
	switch s.state {
	case splashSceneStateInit:
		s.state = splashSceneStateFadeIn
		s.counterMax = ebiten.MaxTPS() / 2
		s.counter = s.counterMax
	case splashSceneStateFadeIn:
		s.counter--
		if s.counter <= 0 {
			s.state = splashSceneStateWait
			s.counterMax = ebiten.MaxTPS() * 2
			s.counter = s.counterMax
		}
	case splashSceneStateWait:
		s.counter--
		if s.counter <= 0 || inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
			s.state = splashSceneStateFadeOut
			s.counterMax = ebiten.MaxTPS() / 2
			s.counter = s.counterMax
		}
	case splashSceneStateFadeOut:
		s.counter--
		if s.counter <= 0 {
			s.state = splashSceneStateQuit
		}
	case splashSceneStateQuit:
		sceneSwitcher.SwitchToTitleScene()
	}
	return nil
}

func (s *SplashScene) Draw(screen *ebiten.Image) {
	if s.state == splashSceneStateInit {
		return
	}

	var alpha float64
	switch s.state {
	case splashSceneStateFadeIn:
		alpha = 1 - float64(s.counter)/float64(s.counterMax)
	case splashSceneStateWait:
		alpha = 1
	case splashSceneStateFadeOut:
		alpha = float64(s.counter) / float64(s.counterMax)
	}
	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, alpha)
	screen.DrawImage(s.splashImg, op)
}
