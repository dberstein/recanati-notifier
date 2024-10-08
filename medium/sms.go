package medium

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/dberstein/recanati-notifier/notification"
)

type SMS struct {
	MediumImpl
	to string
}

func NewSMS(to string) *SMS {
	return &SMS{to: to}
}

func (m SMS) String() string {
	// todo: send sms
	return fmt.Sprintf("- SMS to: %s\n", m.to)
}

func (m SMS) Notify(n *notification.Notification) error {
	fmt.Print(m.String())
	if rand.IntN(100) > 50 {
		return errors.New("SMS error")
	}
	return nil
}
