package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/storage"
)

func init() {
	Router().POST("/hosts/:id/fencers", CreateFencer)
	Router().GET("/hosts/:id/fencers", GetHostFencers)
	Router().PUT("/hosts/:id/fencers/:fid", UpdateFencer)
	Router().DELETE("/hosts/:id/fencers/:fid", DeleteFencer)
}

func CreateFencer(c *gin.Context) {
	var fencer storage.HostFencer
	ParseBody(c, &fencer)

	host := GetHost(c)
	fencer.HostId = host.Id
	if 0 == fencer.Port {
		fencer.Port = 623
	}

	// FIXME: validate before insert into database.

	if err := storage.FencerInsert(&fencer); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusOK, fencer)
	}
}

func GetHostFencers(c *gin.Context) {
	host := GetHost(c)

	fencers, err := storage.FencerGetAll(host.Id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, fencers)
}

func UpdateFencer(c *gin.Context) {
	fencerId := GetId(c, "fid")

	fencer, err := storage.FencerGetById(fencerId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if fencer == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	ParseBody(c, fencer)
	err = storage.FencerUpdate(fencerId, fencer)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, fencer)
	}
}

func DeleteFencer(c *gin.Context) {
	err := storage.FencerDelete(GetId(c, "fid"))
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Data(204, "application/json", make([]byte, 0))
	}
}
