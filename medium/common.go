package medium

import "github.com/dberstein/recanati-notifier/notification"

const MaxRetries int = 3

type MediumStatus int

const (
	StatusSuccess MediumStatus = iota
	StatusPending
	StatusRetry
	StatusError
)

type Medium interface {
	Notify(*notification.Notification) error
	RetryStatus() bool
	SetStatus(MediumStatus)
	String() string
}

type MediumImpl struct {
	status  MediumStatus
	retries int
}

func (m *MediumImpl) SetStatus(s MediumStatus) {
	m.status = s
}

func (m *MediumImpl) RetryStatus() bool {
	if m.retries < MaxRetries {
		m.status = StatusRetry
		m.retries++
		return true
	}
	m.status = StatusError
	return false
}
