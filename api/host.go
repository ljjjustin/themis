package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"themis/database"
)

const (
	HostInitialStatus = "initializing"
)

func init() {
	Router().POST("/hosts", CreateHost)
	Router().GET("/hosts", GetAllHosts)
	Router().GET("/hosts/:id", GetOneHost)
	Router().PUT("/hosts/:id", UpdateHost)
	Router().DELETE("/hosts/:id", DeleteHost)
	Router().POST("/hosts/:id/enable", EnableHost)
	Router().POST("/hosts/:id/disable", DisableHost)
}

func CreateHost(c *gin.Context) {
	var host database.Host

	ParseBody(c, &host)
	host.Status = HostInitialStatus
	host.UpdatedAt = time.Now()
	if err := database.HostInsert(&host); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusCreated, host)
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
		c.JSON(http.StatusAccepted, host)
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

func EnableHost(c *gin.Context) {
	id := GetId(c, "id")

	host, err := database.HostGetById(id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	states, err := database.StateGetAll(host.Id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}
	for _, s := range states {
		s.FailedTimes = 0
		database.StateUpdateFields(s, "failed_times")
	}

	host.Disabled = false
	host.Status = HostInitialStatus
	host.UpdatedAt = time.Now()
	database.HostUpdateFields(host, "status", "disabled", "updated_at")
	c.JSON(http.StatusAccepted, host)
}

func DisableHost(c *gin.Context) {
	id := GetId(c, "id")

	host, err := database.HostGetById(id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	host.Disabled = true
	database.HostUpdateFields(host, "disabled")
	c.JSON(http.StatusAccepted, host)
}
