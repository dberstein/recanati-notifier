package notification

import "fmt"

type NotificationType int

const (
	TypeInformational NotificationType = iota
	TypeAlert
	TypeReminder
)

func (nt NotificationType) String() string {
	return [...]string{"Informational", "Alert", "Reminder"}[nt]
}

func (nt NotificationType) EnumIndex() int {
	return int(nt)
}

type NotificationStatus int

const (
	StatusPending NotificationStatus = iota
	StatusSent
	StatusRetrying
	StatusFailed
)

func (ns NotificationStatus) String() string {
	return [...]string{"Pending", "Sent", "Retrying", "Failed"}[ns]
}

func (ns NotificationStatus) EnumIndex() int {
	return int(ns)
}

type NotificationRequest struct {
	Type    NotificationType `json:"type"`
	Content string           `json:"content"`
}

type Notification struct {
	Type    NotificationType
	Status  NotificationStatus
	Retries int
	Content string
}

func NewNotification(nt NotificationType, content string) *Notification {
	return &Notification{
		Type:    nt,
		Status:  StatusPending,
		Content: content,
	}
}

func (n *Notification) String() string {
	return fmt.Sprintf("%s (%s): %s\n", n.Type, n.Status, n.Content)
}
