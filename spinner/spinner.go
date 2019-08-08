package spinner

import (
	"fmt"
	"time"
)

type Spinner struct {
	Msg     string
	states  []string
	state   int
	ticker  *time.Ticker
	stopped bool
	stop    chan bool
}

func New(msg string) *Spinner {
	spinner := &Spinner{
		Msg:    msg,
		states: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
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
			fmt.Printf("\r%s %s", s.Msg, s.next())
		case <-s.stop:
			fmt.Printf("\r%s %s\n", s.Msg, "✓")
			return
		}
	}
}

func (s *Spinner) erase() {

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
