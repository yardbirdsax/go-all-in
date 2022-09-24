package pomo

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbletea"
)

type Pomo struct {
	duration time.Duration
	timer timer.Model
}

func NewPomo(duration time.Duration) Pomo {
	pomo := Pomo{
		duration: duration,
		timer: timer.New(duration),
	}
	return pomo
}

func (p Pomo) Init() tea.Cmd {
	return p.timer.Init()
}

func (p Pomo) Update (msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case timer.TickMsg:
		p.timer, cmd = p.timer.Update(msg)
		return p, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
			return p, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			p.timer.Stop()
			p.timer.Timeout = p.duration
			cmd = p.timer.Start()
			return p, cmd
		}
	}
	return p, nil
}

func (p Pomo) View() string {
	var s string
	if p.timer.Timedout() {
		s = "All done!"
	} else {
		s = p.timer.Timeout.String()
	}
	s += "\n"
	return s
}
