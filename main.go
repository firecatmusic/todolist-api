package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"todolist/model"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

func main() {
	// Initialize and persist the database connection
	db, err := connect()
	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	router := gin.Default()
	router.GET("/getuser", getListUser)
	router.POST("/register", registerUser)

	router.Run("localhost:8000")
}

func connect() (*sql.DB, error) {
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

func getListUser(c *gin.Context) {
	var filteredUser model.User
	db, err := connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()
	// Get query parameters
	id := c.Query("id")

	if id == "" {
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Parameter id is required!"})
	} else {
		// Ensure the query runs and rows.Close() is deferred
		rows, err := db.Query("SELECT id, name, username, password, token, image FROM user WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": err.Error()})
			return
		}
		defer rows.Close() // Properly defer closing of rows

		// A flag to track if any rows are found
		found := false

		// Iterate through the result set
		for rows.Next() {
			var user model.User
			err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Token, &user.Image)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": err.Error()})
				return
			}
			// Append each user to the slice
			filteredUser = user
			found = true // Mark that we found a row
		}

		// Check for errors during iteration
		if err = rows.Err(); err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": err.Error()})
			return
		}

		if found {
			// Return the response after all rows have been processed
			c.JSON(http.StatusOK, gin.H{"api_status": 1, "api_message": "Success", "data": filteredUser})

		} else {
			// Return the response after all rows have been processed
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "User tidak ditemukan!"})

		}
	}
}

func registerUser(c *gin.Context) {
	// Your registration logic here
}
