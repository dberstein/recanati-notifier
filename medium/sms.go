package medium

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/dberstein/recanati-notifier/notification"
	"github.com/fatih/color"
)

type SMS struct {
	MediumImpl
	to string
}

func NewSMS(to string) *SMS {
	return &SMS{to: to}
}

func (m SMS) String() string {
	return fmt.Sprintf("SMS (%s)\n", m.to)
}

func (m SMS) Notify(n *notification.Message) error {
	fmt.Print(color.YellowString(m.String()))
	time.Sleep(time.Duration(rand.IntN(500) * int(time.Millisecond)))
	if rand.IntN(100) > 100-PctError {
		return errors.New("SMS error")
	}
	return nil
}
