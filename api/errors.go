package api

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotFound         = errors.New("Resource not found.")
	ErrInvalidParameter = errors.New("Invalid parameters.")
	ErrDuplicatedTag    = errors.New("tag must be unique for one host.")
)

type HTTPError struct {
	code int
	msg  string
}

func AbortWithError(code int, err error) {
	panic(&HTTPError{code: code, msg: err.Error()})
}

func FaultWrap() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(*HTTPError); ok {
					c.AbortWithStatusJSON(e.code, gin.H{"error": e.msg})
				} else {
					c.AbortWithStatus(500)
				}
			}
		}()
		c.Next()
	}
}
