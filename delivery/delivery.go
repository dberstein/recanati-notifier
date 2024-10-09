package delivery

import (
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/user"
)

type Delivery struct {
	User    *user.User
	Message *notification.Message
}

func New(u *user.User, msg *notification.Message) *Delivery {
	return &Delivery{User: u, Message: msg}
}
