package queue

import (
	"fmt"
	"sync"

	"github.com/dberstein/recanati-notifier/delivery"
	"github.com/dberstein/recanati-notifier/medium"
	"github.com/dberstein/recanati-notifier/notification"
	"github.com/dberstein/recanati-notifier/user"

	"github.com/fatih/color"
)

var lock sync.RWMutex

type Queue []*delivery.Delivery

func NewQueue() *Queue {
	return &Queue{}
}

func (self *Queue) Push(x *delivery.Delivery) {
	lock.Lock()
	defer lock.Unlock()

	*self = append(*self, x)
}

func (self *Queue) Pop() *delivery.Delivery {
	lock.Lock()
	defer lock.Unlock()

	h := *self
	var el *delivery.Delivery
	if len(h) > 0 {
		el, *self = h[0], h[1:]
	}
	return el
}

func (self *Queue) Notify(u *user.User, msg *notification.Message) {
	for _, m := range u.Mediums {
		d := delivery.New(u, msg)
		var colorPrint = color.GreenString
		var logMsg string
		if err := m.Notify(msg); err != nil {
			colorPrint = color.RedString
			m.SetStatus(medium.StatusRetry)
			if m.Retry() {
				self.Push(d)
			}

			logMsg = err.Error()
		} else {
			logMsg = "success"
			m.SetStatus(medium.StatusSuccess)
		}
		fmt.Println(colorPrint(logMsg))
	}
}
