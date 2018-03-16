package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/storage"
)

func GetId(c *gin.Context, key string) int {
	id, err := strconv.Atoi(c.Param(key))
	if err != nil {
		AbortWithError(http.StatusBadRequest, err)
	}
	return int(id)
}

func GetKey(c *gin.Context, key string, minLen int) string {
	value := c.Param(key)
	if len(value) < minLen {
		AbortWithError(http.StatusBadRequest, ErrInvalidParameter)
	}
	return value
}

func ParseBody(c *gin.Context, obj interface{}) {
	if err := c.Bind(obj); err != nil {
		AbortWithError(http.StatusBadRequest, err)
	}
}

func GetHost(c *gin.Context) *storage.Host {
	hostId := GetId(c, "id")

	host, err := storage.HostGetById(hostId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusBadRequest, ErrNotFound)
	}
	return host
}
