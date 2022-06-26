// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"errors"
)

var TaskEnded = errors.New("task ended")

type Task func() error

type Sequence struct {
	tasks []Task
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

func (s *Sequence) AddTask(f Task) {
	s.tasks = append(s.tasks, f)
}

func NewTimerTask(f func(counter int, maxCounter int) error, counter int) Task {
	var current int
	max := counter
	return func() error {
		current++
		if err := f(current, max); err != nil {
			return err
		}
		if current >= max {
			return TaskEnded
		}
		return nil
	}
}

func NewAllTask(tasks ...Task) Task {
	return func() error {
		var execed bool
		for i, t := range tasks {
			if t == nil {
				continue
			}
			execed = true
			if err := t(); err != nil {
				if err == TaskEnded {
					tasks[i] = nil
					continue
				}
				return err
			}
		}
		if execed {
			return nil
		}
		return TaskEnded
	}
}
