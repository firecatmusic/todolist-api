package main

import (
	apiTasks "todolist/api/tasks"
	apiUser "todolist/api/user"
	model "todolist/model"

	gin "github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.MaxMultipartMemory = 1 // 1 MiB

	//AUTH API
	router.POST("/login", apiUser.Login)
	router.POST("/register", apiUser.RegisterUser)
	router.POST("/upload_image", apiUser.UploadImage)

	//TASK API
	router.POST("/get_list_tasks", apiTasks.GetAllTasks)
	router.POST("/update_status_task", apiTasks.UpdateStatusTask)
	router.POST("/delete_task", apiTasks.DeleteTask)
	router.POST("/create_task", apiTasks.CreateTask)

	router.Run(model.BaseURL)
}
