package rest

import (
	"github.com/rosbit/mgin"
	"delay-tasks/utils"
	"net/http"
	"strings"
	"time"
	"fmt"
)

// POST /task/:cate
// BODY: {"timestamp": <ts>, "key": uint64, params: {anything}, "handler":""}
func AddTask(c *mgin.Context) {
	var params struct {
		Timestamp int64
		Key       uint64
		Params    interface{}
		Handler   string
	}
	cate := c.Param("cate")
	if status, err := c.ReadJSON(&params); err != nil {
		c.Error(status, err.Error())
		return
	}
	now := time.Now().Unix()
	if params.Timestamp < now && params.Timestamp + utils.SLOT_TIME_SPAN < now {
		c.Error(http.StatusBadRequest, fmt.Sprintf("timestamp not spicfied or it is a history time: %d", params.Timestamp))
		return
	}
	if params.Key == 0 {
		c.Error(http.StatusBadRequest, "please specify a non-zero key")
		return
	}
	if len(params.Handler) > 0 && !strings.HasPrefix(params.Handler, "http") {
		c.Error(http.StatusBadRequest, `handler can be empty, or it must start with "http"`)
		return
	}
	taskTime := time.Unix(params.Timestamp, 0)
	utils.AddTask(taskTime, cate, params.Key, params.Params, params.Handler)

	c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg": "OK",
	})
}
