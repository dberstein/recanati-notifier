package medium

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/dberstein/recanati-notifier/notification"
)

type Email struct {
	MediumImpl
	to string
}

func NewEmail(to string) *Email {
	return &Email{to: to}
}

func (m Email) String() string {
	// todo: send email
	return fmt.Sprintf("- EMAIL to: %s\n", m.to)
}

func (m Email) Notify(n *notification.Notification) error {
	fmt.Print(m.String())
	if rand.IntN(100) > 50 {
		return errors.New("Email error")
	}
	return nil
}
