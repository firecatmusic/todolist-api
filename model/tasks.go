package model

import "time"

// Task represents a task in the to-do list
type Task struct {
	IDTask      int       `json:"id_task"`
	IDUser      int       `json:"id_user"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
