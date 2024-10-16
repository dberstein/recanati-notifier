package medium

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/dberstein/recanati-notifier/notification"
	"github.com/fatih/color"
)

type Email struct {
	MediumImpl
	to string
}

func NewEmail(to string) *Email {
	return &Email{to: to}
}

func (m Email) String() string {
	return fmt.Sprintf("EMAIL (%s)\n", m.to)
}

func (m Email) Notify(n *notification.Message) error {
	fmt.Print(color.YellowString(m.String()))
	time.Sleep(time.Duration(rand.IntN(500) * int(time.Millisecond)))
	if rand.IntN(100) > 100-PctError {
		return errors.New("Email error")
	}
	return nil
}
