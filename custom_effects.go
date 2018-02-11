package main

import (
	"github.com/faiface/beep"
)

type inifiteLoop struct {
	s      beep.StreamSeeker
	origin int
}

func (l *inifiteLoop) Stream(samples [][2]float64) (n int, ok bool) {
	if l.s.Err() != nil {
		return 0, false
	}
	for len(samples) > 0 {
		sn, sok := l.s.Stream(samples)
		if !sok {
			err := l.s.Seek(0)
			if err != nil {
				return n, true
			}
			continue
		}
		samples = samples[sn:]
		n += sn
	}
	return n, true
}

func (l *inifiteLoop) Err() error {
	return l.s.Err()
}

func InfiniteLoop(s beep.StreamSeeker) beep.Streamer {
	return &inifiteLoop{
		s:      s,
	}
}
