// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SplashScene struct {
	splashImg   *ebiten.Image
	sequence    *Sequence
	splashAlpha float64
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
	if s.sequence == nil {
		s.sequence = &Sequence{}
		s.sequence.AddTask(NewCountingTask(func(counter, maxCounter int) error {
			s.splashAlpha = float64(counter) / float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()/2))
		s.sequence.AddTask(NewCountingTask(func(counter, maxCounter int) error {
			s.splashAlpha = 1
			if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyN) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				return TaskEnded
			}
			return nil
		}, ebiten.MaxTPS()*2))
		s.sequence.AddTask(NewCountingTask(func(counter, maxCounter int) error {
			s.splashAlpha = 1 - float64(counter)/float64(maxCounter)
			return nil
		}, ebiten.MaxTPS()/2))
		s.sequence.AddTask(func() error {
			s.splashAlpha = 0
			sceneSwitcher.SwitchToGameScene()
			return TaskEndedAndContinue
		})
	}
	if err := s.sequence.Update(); err != nil {
		return err
	}
	return nil
}

func (s *SplashScene) Draw(screen *ebiten.Image) {
	if s.sequence == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(1, 1, 1, s.splashAlpha)
	screen.DrawImage(s.splashImg, op)
}
