package pomo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
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