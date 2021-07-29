package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

const (
	DB_SDN = "postgres://lisibo:123456@127.0.0.1:5432/demo?sslmode=disable"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Relationship struct {
	// OwnerID int    `json:"owner_id"`
	UserID int    `json:"user_id"`
	State  string `json:"state"`
	Type   string `json:"type"`
}

func main() {
	// build DB connection
	db, err := sql.Open("postgres", DB_SDN)
	if err != nil {
		log.Fatal("Failed to connect Pg: ", err)
	}
	defer db.Close()
	// gin
	router := gin.Default()

	// show all users
	router.GET("/users", func(c *gin.Context) {

		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		user := make([]User, 0, 10)

		for rows.Next() {
			us := User{}
			err := rows.Scan(&us.ID, &us.Name, &us.Type)
			if err != nil {
				log.Fatal(err)
			}
			user = append(user, us)
		}

		c.IndentedJSON(http.StatusOK, user)
	})

	// add a user
	router.POST("/users", func(c *gin.Context) {
		var us = User{
			Type: "user",
		}
		if err := c.ShouldBindJSON(&us); err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)

		}
		_, err := db.Exec("INSERT INTO users (name, type) VALUES ($1, $2)", us.Name, us.Type)
		if err != nil {
			log.Fatal(err)
		}

		err = db.QueryRow("SELECT MAX(id) FROM users WHERE name = $1", us.Name).Scan(&us.ID)
		if err != nil {
			log.Fatal(err)
		}

		c.IndentedJSON(http.StatusOK, us)
	})

	router.GET("/users/:user_id/relationships", func(c *gin.Context) {
		ownerId := c.Param("user_id")
		rows, err := db.Query("SELECT user_id, state, type FROM relationships WHERE owner_id = $1", ownerId)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		relat := make([]Relationship, 0, 10)

		for rows.Next() {
			rl := Relationship{}
			err := rows.Scan(&rl.UserID, &rl.State, &rl.Type)
			if err != nil {
				log.Fatal(err)
			}
			relat = append(relat, rl)
		}

		c.IndentedJSON(http.StatusOK, relat)

	})

	router.PUT("/users/:user_id/relationships/:other_user_id", func(c *gin.Context) {
		ownerId, _ := strconv.Atoi(c.Param("user_id"))
		userId, _ := strconv.Atoi(c.Param("other_user_id"))

		var rls = Relationship{
			UserID: userId,
			Type:   "relationship",
		}
		if err := c.ShouldBindJSON(&rls); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		}

		var curState string
		err = db.QueryRow(
			"SELECT state FROM relationships WHERE owner_id = $1 AND user_id = $2",
			ownerId, userId).Scan(&curState)

		if curState == "" {
			if rls.State == "liked" {
				var oppoState string
				err = db.QueryRow("SELECT state FROM relationships WHERE user_id = $1 AND owner_id = $2",
					ownerId, userId).Scan(&oppoState)

				if oppoState == "liked" {
					rls.State = "matched"
					oppoState = "matched"
					_, err = db.Exec(
						"UPDATE relationships SET state = $1 WHERE user_id = $2 AND owner_id = $3",
						oppoState, ownerId, userId)

					_, err = db.Exec(
						"INSERT INTO relationships VALUES ($1, $2, $3, 'relationship')",
						ownerId, userId, rls.State)

				} else {
					_, err = db.Exec(
						"INSERT INTO relationships VALUES ($1, $2, $3, 'relationship')",
						ownerId, userId, rls.State)
				}
			} else {
				_, err = db.Exec(
					"INSERT INTO relationships VALUES ($1, $2, $3, 'relationship')",
					ownerId, userId, rls.State)

			}
		} else if curState != rls.State {
			if (curState == "liked" && rls.State == "disliked") || (curState == "disliked" && rls.State == "liked") {
				_, err = db.Exec(
					"UPDATE relationships SET state = $1 WHERE owner_id = $2 AND user_id = $3",
					rls.State, ownerId, userId)

			} else if rls.State == "disliked" && curState == "matched" {
				_, err = db.Exec(
					"UPDATE relationships SET state = $1 WHERE owner_id = $2 AND user_id = $3",
					rls.State, ownerId, userId)

				_, err = db.Exec(
					"UPDATE relationships SET state = 'liked' WHERE user_id = $1 AND owner_id = $2",
					ownerId, userId)
			}
		}

		c.IndentedJSON(http.StatusOK, rls)

	})

	router.Run(":80") // listen and serve on 0.0.0.0:8080
}
