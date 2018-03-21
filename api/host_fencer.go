package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/database"
)

func init() {
	Router().POST("/hosts/:id/fencers", CreateFencer)
	Router().GET("/hosts/:id/fencers", GetHostFencers)
	Router().PUT("/hosts/:id/fencers/:fid", UpdateFencer)
	Router().DELETE("/hosts/:id/fencers/:fid", DeleteFencer)
}

func CreateFencer(c *gin.Context) {
	var fencer database.HostFencer
	ParseBody(c, &fencer)

	host := GetHost(c)
	fencer.HostId = host.Id
	if 0 == fencer.Port {
		fencer.Port = 623
	}

	// FIXME: validate before insert into database.

	if err := database.FencerInsert(&fencer); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusOK, fencer)
	}
}

func GetHostFencers(c *gin.Context) {
	host := GetHost(c)

	fencers, err := database.FencerGetAll(host.Id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, fencers)
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
		c.JSON(200, fencer)
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
