package medium

import (
	"github.com/dberstein/recanati-notifier/notification"
)

const PctError int = 15
const MaxRetries int = 3

type MediumStatus int

const (
	StatusPending MediumStatus = iota
	StatusSuccess
	StatusRetry
	StatusError
)

type Medium interface {
	Notify(*notification.Message) error
	Retry() bool
	SetStatus(MediumStatus)
	GetStatus() MediumStatus
	String() string
}

type MediumImpl struct {
	status  MediumStatus
	retried int
}

func (m *MediumImpl) SetStatus(s MediumStatus) {
	m.status = s
}

func (m *MediumImpl) GetStatus() MediumStatus {
	return m.status
}

func (m *MediumImpl) Retry() bool {
	m.retried++
	if m.retried < MaxRetries {
		m.status = StatusRetry
		return true
	}
	m.status = StatusError
	return false
}
