package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	DB_SDN = "postgres://lisibo:123456@127.0.0.1:5432/demo?sslmode=disable"
)

// 代码最好分层，不要写在一个文件/方法中，可参考ttx-core

func main() {
	// gin
	router := gin.Default()

	// show all users
	router.GET("/users", showAllUsers)
	// add a user
	router.POST("/users", addNewUser)
	// show relationships
	router.GET("/users/:user_id/relationships", showRelationships)
	// add new relationships
	router.PUT("/users/:user_id/relationships/:other_user_id", addNewRelationship)

	router.Run(":80") // listen and serve on 0.0.0.0:8080
}
