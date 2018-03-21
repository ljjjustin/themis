package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/database"
)

func init() {
	Router().POST("/hosts", CreateHost)
	Router().GET("/hosts", GetAllHosts)
	Router().GET("/hosts/:id", GetOneHost)
	Router().PUT("/hosts/:id", UpdateHost)
	Router().DELETE("/hosts/:id", DeleteHost)
}

func CreateHost(c *gin.Context) {
	var host database.Host

	ParseBody(c, &host)
	if err := database.HostInsert(&host); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusOK, host)
	}
}

func GetOneHost(c *gin.Context) {
	id := GetId(c, "id")

	if host, err := database.HostGetById(id); err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusNotFound, ErrNotFound)
	} else {
		c.JSON(http.StatusOK, host)
	}
}

func GetAllHosts(c *gin.Context) {
	hosts, err := database.HostGetAll()
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, hosts)
}

func UpdateHost(c *gin.Context) {
	id := GetId(c, "id")

	host, err := database.HostGetById(id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	ParseBody(c, host)
	if err := database.HostUpdate(id, host); err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, host)
	}
}

func DeleteHost(c *gin.Context) {
	id := GetId(c, "id")

	if err := database.HostDelete(id); err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Data(204, "application/json", make([]byte, 0))
	}
}
