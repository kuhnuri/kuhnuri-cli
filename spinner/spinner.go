package spinner

import (
	"fmt"
	"strings"
	"time"
)

type Spinner struct {
	msg     string
	clean   string
	states  []string
	state   int
	ticker  *time.Ticker
	stopped bool
	stop    chan bool
}

func New(msg string) *Spinner {
	spinner := &Spinner{
		msg:    msg,
		states: []string{"⣷", "⣯", "⣟", "⡿", "⢿", "⣻", "⣽", "⣾"},
		state:  -1,
		ticker: time.NewTicker(time.Millisecond * 100),
		stop:   make(chan bool, 1),
	}
	go spinner.run()
	return spinner
}

func (s *Spinner) run() {
	defer s.ticker.Stop()
	for {
		select {
		case <-s.ticker.C:
			fmt.Printf("\r%s %s%s", s.msg, s.next(), s.clean)
		case <-s.stop:
			fmt.Printf("\r%s %s%s\n", s.msg, "✓", s.clean)
			return
		}
	}
}

func (s *Spinner) Message(msg string) {
	diff := len(s.msg) - len(msg)
	if diff > 0 {
		s.clean = strings.Repeat(" ", diff)
	} else {
		s.clean = ""
	}
	s.msg = msg
}

func (s *Spinner) Stop() {
	close(s.stop)
}

func (s *Spinner) next() string {
	i := s.state + 1
	if i >= len(s.states) {
		s.state = 0
	} else {
		s.state = i
	}
	return s.states[s.state]
}
