package main

import (
	"database/sql"
	"dtest/main/errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type Relationship struct {
	// OwnerID int    `json:"owner_id"`
	UserID int    `json:"user_id"`
	State  string `json:"state"`
	Type   string `json:"type"`
}

const (
	Matched  = "matched"
	Liked    = "liked"
	Disliked = "disliked"
)

var showRelationships = func(c *gin.Context) {
	// build DB connection
	db, err := sql.Open("postgres", DB_SDN)
	if err != nil {
		log.Fatal("Failed to connect Pg: ", err)
	}
	defer db.Close()

	ownerId := c.Param("user_id")
	rows, queryErr := db.Query("SELECT user_id, state, type FROM relationships WHERE owner_id = $1", ownerId)
	if queryErr != nil {
		c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
		return
	}
	defer rows.Close()

	relat := make([]Relationship, 0, 10)

	for rows.Next() {
		rl := Relationship{}
		err := rows.Scan(&rl.UserID, &rl.State, &rl.Type)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, errors.QueryError)
			return
		}
		relat = append(relat, rl)
	}

	c.IndentedJSON(http.StatusOK, relat)

}

var addNewRelationship = func(c *gin.Context) {
	// build DB connection
	db, err := sql.Open("postgres", DB_SDN)
	if err != nil {
		log.Fatal("Failed to connect Pg: ", err)
	}
	defer db.Close()

	ownerId, _ := strconv.Atoi(c.Param("user_id"))
	userId, _ := strconv.Atoi(c.Param("other_user_id"))

	var rls = Relationship{
		UserID: userId,
		Type:   "relationship",
	}
	if err := c.ShouldBindJSON(&rls); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
	}
	// 参数校验
	var curState string
	err = db.QueryRow(
		"SELECT state FROM relationships WHERE owner_id = $1 AND user_id = $2",
		ownerId, userId).Scan(&curState)

	if curState == "" {
		if rls.State == Liked {
			var oppoState string
			err = db.QueryRow("SELECT state FROM relationships WHERE user_id = $1 AND owner_id = $2",
				ownerId, userId).Scan(&oppoState)

			if oppoState == Liked {
				// 状态值最好定义const常量
				rls.State = Matched
				oppoState = Matched
				// 两个更新语句需要使用事务，且select最新状态时，需要加锁
				_, err = db.Exec(
					"UPDATE relationships SET state = $1 WHERE user_id = $2 AND owner_id = $3",
					oppoState, ownerId, userId)

				_, err = db.Exec(
					"INSERT INTO relationships VALUES ($1, $2, $3, 'relationship')", //
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
		if (curState == Liked && rls.State == Disliked) || (curState == Disliked && rls.State == Liked) {
			_, err = db.Exec(
				"UPDATE relationships SET state = $1 WHERE owner_id = $2 AND user_id = $3",
				rls.State, ownerId, userId)

		} else if rls.State == Disliked && curState == Matched {
			// 事务
			_, err = db.Exec(
				"UPDATE relationships SET state = $1 WHERE owner_id = $2 AND user_id = $3",
				rls.State, ownerId, userId)

			_, err = db.Exec(
				"UPDATE relationships SET state = 'liked' WHERE user_id = $1 AND owner_id = $2",
				ownerId, userId)
		}
	}
	// error未处理

	c.IndentedJSON(http.StatusOK, rls)

}
