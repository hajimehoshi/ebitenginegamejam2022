// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

package main

import (
	"errors"
)

var (
	TaskEnded            = errors.New("task ended")
	TaskEndedAndContinue = errors.New("task ended and continue")
)

type Task func() error

type Sequence struct {
	tasks []Task
}

func (s *Sequence) Update() error {
retry:
	if len(s.tasks) == 0 {
		return nil
	}
	if err := s.tasks[0](); err != nil {
		if err == TaskEnded || err == TaskEndedAndContinue {
			s.tasks[0] = nil
			s.tasks = s.tasks[1:]
			if err == TaskEndedAndContinue {
				goto retry
			}
			return nil
		}
		return err
	}
	return nil
}

func (s *Sequence) AddTask(f Task) {
	s.tasks = append(s.tasks, f)
}

func NewCountingTask(f func(counter int, maxCounter int) error, counter int) Task {
	var current int
	max := counter
	return func() error {
		if err := f(current, max); err != nil {
			return err
		}
		current++
		if current >= max {
			return TaskEnded
		}
		return nil
	}
}

func NewAllTask(tasks ...Task) Task {
	cont := true
	return func() error {
		var execed bool
		for i, t := range tasks {
			if t == nil {
				continue
			}
			execed = true
			if err := t(); err != nil {
				if err == TaskEnded || err == TaskEndedAndContinue {
					tasks[i] = nil
					if err == TaskEnded {
						cont = false
					}
					continue
				}
				return err
			}
		}
		if execed {
			return nil
		}
		if cont {
			return TaskEndedAndContinue
		}
		return TaskEnded
	}
}
