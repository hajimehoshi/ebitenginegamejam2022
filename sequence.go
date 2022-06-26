// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"errors"
)

var TaskEnded = errors.New("task ended")

type Sequence struct {
	tasks []func() error
}

func (s *Sequence) Update() error {
	if len(s.tasks) == 0 {
		return nil
	}
	if err := s.tasks[0](); err != nil {
		if err == TaskEnded {
			s.tasks[0] = nil
			s.tasks = s.tasks[1:]
			return nil
		}
		return err
	}
	return nil
}

func (s *Sequence) AddTask(f func() error) {
	s.tasks = append(s.tasks, f)
}

func (s *Sequence) AddTimerTask(f func(counter int, maxCounter int) error, counter int) {
	var current int
	max := counter
	s.AddTask(func() error {
		current++
		if err := f(current, max); err != nil {
			return err
		}
		if current >= max {
			return TaskEnded
		}
		return nil
	})
}
