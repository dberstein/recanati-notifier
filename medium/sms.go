package medium

import (
	"errors"
	"fmt"
	"math/rand/v2"

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
	if rand.IntN(100) > 100-PctSuccess {
		return errors.New("SMS error")
	}
	return nil
}
