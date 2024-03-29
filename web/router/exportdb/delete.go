package exportdb

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mybee/go-spider/web/core"
	"github.com/mybee/go-spider/web/model"
	log "github.com/sirupsen/logrus"
)

func DeleteExportDB(c *gin.Context) {
	idStr := c.Param("id")
	log.Infof("delete exportdb id [%v]", idStr)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	edb := &model.ExportDB{ID: id}
	if err := edb.Delete(core.GetGormDB()); err != nil {
		log.Errorf("delete export db err [%+v]", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusNoContent, "", nil)
}
