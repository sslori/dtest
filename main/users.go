package main

import (
	"database/sql"
	"dtest/main/errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

var showAllUsers = func(c *gin.Context) {
	// build DB connection
	db, err := sql.Open("postgres", DB_SDN)
	if err != nil {
		log.Fatal("Failed to connect Pg: ", err)
	}
	defer db.Close()

	rows, queryErr := db.Query("SELECT * FROM users")
	if queryErr != nil {
		c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
		return
	}
	defer rows.Close()

	user := make([]User, 0, 10)

	for rows.Next() {
		us := User{}
		err := rows.Scan(&us.ID, &us.Name, &us.Type)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
			return
		}
		user = append(user, us)
	}
	c.IndentedJSON(http.StatusOK, user) // 出现error的情况的返回
}

var addNewUser = func(c *gin.Context) {
	// build DB connection
	db, err := sql.Open("postgres", DB_SDN)
	if err != nil {
		log.Fatal("Failed to connect Pg: ", err)
	}
	defer db.Close()

	var us = User{
		Type: "user",
	}
	if err := c.ShouldBindJSON(&us); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	// 接口需要做一些参数校验
	if us.Name == "" || len(us.Name) > 20 {
		c.IndentedJSON(http.StatusInternalServerError, errors.InvalidName)
		return
	}

	_, err = db.Exec("INSERT INTO users (name, type) VALUES ($1, $2)", us.Name, us.Type)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
		return
	}
	// 并发情况下可能有问题，上面insert语句可以直接返回id
	err = db.QueryRow("SELECT MAX(id) FROM users WHERE name = $1", us.Name).Scan(&us.ID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
		return
	}

	c.IndentedJSON(http.StatusOK, us)

}
