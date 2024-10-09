package user

import (
	"github.com/dberstein/recanati-notifier/medium"
)

type User struct {
	Id      int
	Mediums []medium.Medium
}
