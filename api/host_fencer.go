package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/database"
)

func init() {
	Router().GET("/fencers", ListFencers)
	Router().GET("/fencers/:fid", GetFencer)
	Router().POST("/fencers", CreateFencer)
	Router().PUT("/fencers/:fid", UpdateFencer)
	Router().DELETE("/fencers/:fid", DeleteFencer)
	Router().GET("/fencedTimes", ShowFencedTimes)
}

func ListFencers(c *gin.Context) {
	fencers, err := database.FencerGetAll()
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, fencers)
}

func GetFencer(c *gin.Context) {
	fencerId := GetId(c, "fid")

	fencer, err := database.FencerGetById(fencerId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if fencer == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	c.JSON(http.StatusOK, fencer)
}

func CreateFencer(c *gin.Context) {
	var fencer database.HostFencer
	ParseBody(c, &fencer)

	host, err := database.HostGetById(fencer.HostId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusBadRequest, ErrNotFound)
	}
	if 0 == fencer.Port {
		fencer.Port = 623
	}

	// FIXME: validate before insert into database.

	if err := database.FencerInsert(&fencer); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusCreated, fencer)
	}
}

func UpdateFencer(c *gin.Context) {
	fencerId := GetId(c, "fid")

	fencer, err := database.FencerGetById(fencerId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if fencer == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	ParseBody(c, fencer)
	err = database.FencerUpdate(fencerId, fencer)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusAccepted, fencer)
	}
}

func DeleteFencer(c *gin.Context) {
	err := database.FencerDelete(GetId(c, "fid"))
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Data(204, "application/json", make([]byte, 0))
	}
}

func ShowFencedTimes(c *gin.Context) {

	fencedTimes, err := database.GetFencedTimes()
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {

		resp := struct {
			FencedTimes int64 `json:"fenced_times"`
		}{
			fencedTimes,
		}

		c.JSON(http.StatusOK, resp)
	}
}