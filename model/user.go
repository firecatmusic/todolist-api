package model

import (
	"todolist/helper"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID       int               `json:"id"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Name     helper.NullString `json:"name"`
	Token    helper.NullString `json:"token"`
	Image    helper.NullString `json:"image"`
}
