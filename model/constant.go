package model

const (
	DatabaseName   = "todolist_db"
	DataSourceName = "root:@tcp(localhost:8868)/todolist_db?timeout=60s"
	// API related constants
	BaseURL         = "localhost:8000"
	TimeoutDuration = 60 // in seconds
)
