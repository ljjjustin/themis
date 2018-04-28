package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/database"
)

func init() {

	Router().GET("/leader", GetLeader)
}

func GetLeader(c *gin.Context) {

	leader, err := database.GetLeader("themisLeader")
	if err != nil {
		AbortWithError(http.StatusNotFound, err)
	}

	c.JSON(http.StatusOK, leader)
}