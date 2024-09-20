package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
	model "todolist/model"

	"github.com/gin-gonic/gin"
)

// Directory to save uploaded images
const uploadDir = "./uploads"

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

func Login(c *gin.Context) {
	db, err := Connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	// Get query parameters
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Parameter username / password is required!"})
	} else {
		// Ensure the query runs and rows.Close() is deferred

		query := "SELECT id, name, username, password, token, image FROM user WHERE username = ? AND password = ?"
		var user model.User

		// Execute the query and scan the result into the user struct
		err := db.QueryRowContext(context.Background(), query, username, password).Scan(
			&user.ID, &user.Name, &user.Username, &user.Password, &user.Token, &user.Image,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Username / Password salah!"})
			} else {
				c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": err.Error()})
			}
			return
		}

		// Return the response after all rows have been processed
		c.JSON(http.StatusOK, gin.H{"api_status": 1, "api_message": "Success", "data": user})
	}
}

func RegisterUser(c *gin.Context) {
	// Your registration logic here

	db, errDB := Connect()

	if errDB != nil {
		fmt.Println("Database connection failed:", errDB)
		defer db.Close()
		return
	} else {
		bodyBytes, errJson := io.ReadAll(c.Request.Body)
		if errJson != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read request body"})
			return
		}

		var user model.User

		err := json.Unmarshal(bodyBytes, &user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
			return
		}

		log.Println("Raw Body: ", user)

		// Since the body has already been read, we need to set it back to be used later
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Continue to the next handler
		c.Next()

		query := `INSERT INTO user (username, name, password) VALUES (?, ?, ?)`
		insertResult, err := db.ExecContext(context.Background(), query, user.Username, user.Name.String, user.Password)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": err.Error()})
			return
		}

		// You can check the result, such as the last inserted ID or the number of rows affected
		lastInsertID, err := insertResult.LastInsertId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to retrieve last insert ID"})
			return
		}

		rowsAffected, err := insertResult.RowsAffected()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to retrieve rows affected"})
			return
		}

		fmt.Println(`lastInsertedId :`, lastInsertID, rowsAffected)

		// Respond with the result
		c.JSON(http.StatusOK, gin.H{
			"api_status":  1,
			"api_message": "User successfully created",
		})
	}
}

// SanitizeFilename replaces special characters and spaces in the filename
func SanitizeFilename(filename string) string {
	// Replace spaces with underscores
	sanitized := regexp.MustCompile(`\s+`).ReplaceAllString(filename, "_")

	// Replace special characters (keeping . and _) with an empty string
	sanitized = regexp.MustCompile(`[^\w\.-]`).ReplaceAllString(sanitized, "")

	return sanitized
}

// UploadImage handles image upload
func UploadImage(c *gin.Context) {
	idUser := c.PostForm("id_user")
	fmt.Println(idUser)

	// Get the file from the request (in Postman, set the key as "file")
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"api_status": 0, "api_message": "Failed to upload image: " + err.Error()})
		return
	}
	defer file.Close()

	// Create the uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"api_status": 0, "api_message": "Unable to create directory"})
		return
	}

	// Get the filename
	filename := GenerateFilename(header.Filename)

	// Create the full path to save the image
	filePath := filepath.Join(uploadDir, SanitizeFilename(filename))

	// Save the file to the uploads directory
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"api_status": 0, "api_message": "Unable to save the file"})
		return
	}
	defer out.Close()

	// Copy the uploaded file to the created file
	if _, err := out.ReadFrom(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"api_status": 0, "api_message": "Failed to save file"})
		return
	}

	db, err := Connect()

	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	defer db.Close()

	if idUser != "" {
		// Ensure the query runs and rows.Close() is deferred

		query := "UPDATE user SET image = ? WHERE id = ?"
		selectQuery := "SELECT id, name, username, password, token, image FROM user WHERE id = ?"

		// Execute the query and scan the result into the user struct
		_, err = db.Exec(query, filePath, idUser)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to update user image: " + err.Error()})
			return
		}

		var user model.User
		err = db.QueryRow(selectQuery, idUser).Scan(
			&user.ID, &user.Name, &user.Username, &user.Password, &user.Token, &user.Image,
		)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Failed to retrieve user data: " + err.Error()})
			return
		}

		// Return the response after all rows have been processed
		c.JSON(http.StatusOK, gin.H{"api_status": 1, "api_message": "Success", "data": user})
	} else {
		// Return the response after all rows have been processed
		c.JSON(http.StatusOK, gin.H{"api_status": 0, "api_message": "Field must not empty!"})
	}
}

// RetrieveImage serves an uploaded image
func RetrieveImage(c *gin.Context) {
	filename := c.Param("filename")

	// Generate the file path
	filePath := filepath.Join(uploadDir, filename)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"api_status": 0, "api_message": "File not found"})
		return
	}

	// Serve the file as a response
	c.File(filePath)
}

// GenerateFilename generates a unique filename for the uploaded image
func GenerateFilename(originalFilename string) string {
	sanitized := SanitizeFilename(originalFilename)
	timestamp := time.Now().Unix()                           // Get current timestamp
	extension := filepath.Ext(sanitized)                     // Get the file extension
	baseName := sanitized[0 : len(sanitized)-len(extension)] // Remove extension for the base name

	// Generate the new filename
	return fmt.Sprintf("%s_%d%s", baseName, timestamp, extension)
}
