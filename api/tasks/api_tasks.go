package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"todolist/model"

	"github.com/gin-gonic/gin"
)

func Connect() (*sql.DB, error) {
	db, err := sql.Open("mysql", model.DataSourceName)
	if err != nil {
		return db, fmt.Errorf("error opening the database: %w", err)
	}

	// Check the connection by pinging the database
	err = db.Ping()
	if err != nil {
		return db, fmt.Errorf("error connecting to the database: %w", err)
	}

	fmt.Println("Database connection successful!")
	return db, nil
}

func GetAllTasks(c *gin.Context) {
	var createdAt, updatedAt string

	userID := c.PostForm("id_user")

	println(userID)
	db, errDb := Connect()
	if errDb != nil {
		fmt.Printf("error opening the database: %w", errDb)
		return
	}
	defer db.Close()

	// Query to get all tasks for the given user ID
	query := `SELECT id_task, id_user, title, description, completed, created_at, updated_at FROM tasks WHERE id_user = ?`
	rows, err := db.QueryContext(context.Background(), query, userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to fetch tasks: " + err.Error()})
		return
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task
		err := rows.Scan(&task.IDTask, &task.IDUser, &task.Title, &task.Description, &task.Completed, &createdAt, &updatedAt)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Error scanning task data: " + err.Error()})
			return
		}

		// Parse the date strings into time.Time
		task.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Error parsing created_at: " + err.Error()})
			return
		}

		task.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Error parsing updated_at: " + err.Error()})
			return
		}

		tasks = append(tasks, task)
	}

	// Return the tasks in the response
	c.JSON(http.StatusOK, gin.H{"api_status": 1, "api_message": "Success", "data": tasks})
}

func UpdateStatusTask(c *gin.Context) {

}

func DeleteTask(c *gin.Context) {

}

func CreateTask(c *gin.Context) {

}
