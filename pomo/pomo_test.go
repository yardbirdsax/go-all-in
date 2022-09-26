package pomo

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewPomo(t *testing.T) {
	Convey("TestNewPomo", t, func() {
		expectedDuration := 10 * time.Second

		pomo := NewPomo(expectedDuration)

		Convey("sets the correct duration", func() {
			So(pomo.timer.Timeout, ShouldEqual, expectedDuration)
		})
	})
}

func TestUpdate(t *testing.T) {
	Convey("Update", t, func() {
		initialTimeout := 10 * time.Second
		tests := []struct{
			name string
			msg tea.Msg
			expectedTimeout time.Duration
			expectedCommand tea.Cmd
		}{
			{
				name: "TickMessage",
				msg: timer.TickMsg{
					ID: 2,
					Timeout: false,
				},
				expectedTimeout: initialTimeout - time.Second,
			},
			{
				name: "KeyMessage-q",
				msg: tea.KeyMsg{
					Runes: []rune{'q'},
					Type: tea.KeyRunes,
				},
				expectedTimeout: initialTimeout,
				expectedCommand: tea.Quit,
			},
		}

		for _, test := range tests {
			Convey(test.name, func() {
				timer := timer.New(initialTimeout)
				pomo := Pomo{
					timer: timer,
				}
				pomo.Init()

				model, acutalCommand := pomo.Update(test.msg)
				actualPomo, ok := model.(Pomo)

				So(ok, ShouldBeTrue)
				So(actualPomo.timer.Timeout, ShouldEqual, test.expectedTimeout)
				if (test.expectedCommand != nil) {
					So(acutalCommand, ShouldEqual, test.expectedCommand)
				}
			})
		}
	})
}