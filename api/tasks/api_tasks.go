package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
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

func EditTask(c *gin.Context) {
	db, err := Connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	idTask := c.PostForm("id_task")

	if idTask != "" {

		task, err := getTaskByID(db, idTask)
		if err != nil {
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Task tidak ditemukan!",
			})
			return
		}

		title := c.PostForm("title")
		description := c.PostForm("description")

		if title != "" {
			task.Title = title
		}

		if description != "" {
			task.Description = description
		}

		log.Println(task)

		query := "UPDATE tasks SET title = ?, description = ?, updated_at = NOW() WHERE id_task = ?"
		a, err := db.Exec(query, task.Title, task.Description, task.IDTask)

		if err != nil {
			log.Fatal("Failed to execute DELETE query:", err)
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Failed to execute DELETE query: " + err.Error(),
			})
			return
		}

		// Check how many rows were affecteds
		rowsAffected, err := a.RowsAffected()
		if err != nil {
			log.Fatal("Error fetching affected rows:", err)
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Error fetching affected rows: " + err.Error(),
			})
			return
		}

		fmt.Printf("Rows affected: %d\n", rowsAffected)
		// Respond with the result
		c.JSON(http.StatusOK, gin.H{
			"api_status":  1,
			"api_message": "Berhasil mengedit task!",
		})

	} else {
		// Respond with the result
		c.JSON(http.StatusOK, gin.H{
			"api_status":  0,
			"api_message": "Field kosong!",
		})
	}

}

func DeleteTask(c *gin.Context) {
	db, err := Connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	idTask := c.PostForm("id_task")

	if idTask != "" {

		task, err := getTaskByID(db, idTask)
		if err != nil {
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Task tidak ditemukan!",
			})
			return
		}

		query := "DELETE FROM `tasks` WHERE `tasks`.`id_task` = ?"
		a, err := db.Exec(query, task.IDTask)

		if err != nil {
			log.Fatal("Failed to execute DELETE query:", err)
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Failed to execute DELETE query: " + err.Error(),
			})
			return
		}

		// Check how many rows were affecteds
		rowsAffected, err := a.RowsAffected()
		if err != nil {
			log.Fatal("Error fetching affected rows:", err)
			// Respond with the result
			c.JSON(http.StatusOK, gin.H{
				"api_status":  0,
				"api_message": "Error fetching affected rows: " + err.Error(),
			})
			return
		}

		fmt.Printf("Rows affected: %d\n", rowsAffected)
		// Respond with the result
		c.JSON(http.StatusOK, gin.H{
			"api_status":  1,
			"api_message": "Berhasil menghapus task!",
		})

	} else {
		// Respond with the result
		c.JSON(http.StatusOK, gin.H{
			"api_status":  0,
			"api_message": "Field kosong!",
		})
	}

}

func CreateTask(c *gin.Context) {
	// {
	// 	"id_task": 1,
	// 	"id_user": 0,
	// 	"title": "Sample Task",
	// 	"description": "This is a sample task description.",
	// 	"completed": false,
	// 	"created_at": "2024-09-20T13:43:54Z",
	// 	"updated_at": "2024-09-20T13:43:54Z"
	// }

	db, err := Connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	var requestBody model.TaskRequest

	// Check if body exists before trying to bind JSON
	if err = c.BindJSON(&requestBody); err != nil {
		// Check for an EOF error
		if err == io.EOF {
			fmt.Println("error bindJson: Empty or malformed JSON body")
			// c.JSON(http.StatusBadRequest, gin.H{"api_status": 0, "api_message": "Empty or malformed JSON body"})
			return
		}
		fmt.Printf("error bindJson: %v\n", err)
		// c.JSON(http.StatusBadRequest, gin.H{"api_status": 0, "api_message": err.Error()})
		return
	}

	queryInsertTask := `INSERT INTO tasks (title, description, id_user) VALUES (?, ?, ?)`
	insertResult, errQuery := db.ExecContext(context.Background(), queryInsertTask, requestBody.Title, requestBody.Description, requestBody.IDUser)

	if errQuery != nil {
		fmt.Printf("error insertResult: %w", errQuery)
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": errQuery.Error()})

	}

	// You can check the result, such as the last inserted ID or the number of rows affected
	lastInsertID, errlastInsertID := insertResult.LastInsertId()
	if err != nil {
		fmt.Printf("error lastInsertID: %w", errlastInsertID)
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to create task: " + errlastInsertID.Error()})
	}

	rowsAffected, errrowsAffected := insertResult.RowsAffected()
	if errrowsAffected != nil {
		fmt.Printf("error lastInsertID: %w", errlastInsertID)
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to retrieve rows affected (" + errrowsAffected.Error() + ")"})
	}

	fmt.Println(`lastInsertedId :`, lastInsertID, rowsAffected)

	// Respond with the result
	c.JSON(http.StatusOK, gin.H{
		"api_status":  1,
		"api_message": "Berhasil membuat task!",
	})

}

func getTaskByID(db *sql.DB, id string) (model.Task, error) {
	var task model.Task

	query := "SELECT id_task,title,description FROM tasks WHERE id_task = ?"
	err := db.QueryRow(query, id).Scan(&task.IDTask, &task.Title, &task.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle no results
			fmt.Println("No task found with the given id")
		}

		log.Println(task)
		return task, err
	}

	log.Println(task)
	return task, nil
}
