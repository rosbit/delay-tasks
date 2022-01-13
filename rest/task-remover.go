package rest

import (
	"github.com/rosbit/mgin"
	"delay-tasks/utils"
	"net/http"
	"time"
)

// DELETE /task/:cate
// BODY: {"timestamp": <ts>, "key": uint64, "exec": true|false}
func DelTask(c *mgin.Context) {
	var params struct {
		Timestamp int64
		Key       uint64
		Exec      bool
	}
	cate := c.Param("cate")
	if status, err := c.ReadJSON(&params); err != nil {
		c.Error(status, err.Error())
		return
	}
	if params.Key == 0 {
		c.Error(http.StatusBadRequest, "please specify a non-zero key")
		return
	}
	taskTime := time.Unix(params.Timestamp, 0)
	utils.DelTask(taskTime, cate, params.Key, params.Exec)

	c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg": "OK",
	})
}
