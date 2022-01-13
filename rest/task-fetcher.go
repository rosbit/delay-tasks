package rest

import (
	"github.com/rosbit/mgin"
	"delay-tasks/utils"
	"net/http"
)

// GET /task/:cate/:key[?timestamp=xxx]
func FetchTask(c *mgin.Context) {
	var p struct {
		Cate string `path:"cate"`
		Key  uint64 `path:"key"`
		Timestamp int64 `query:"timestamp" optional:"true"`
	}
	if code, err := c.ReadParams(&p); err != nil {
		c.Error(code, err.Error())
		return
	}
	timestamp := func()int64 {
		if p.Timestamp == 0 {
			return -1
		}
		return p.Timestamp
	}()

	timeToRun, params, handler, ok := utils.FindTask(p.Cate, p.Key, timestamp)
	if !ok {
		c.Error(http.StatusNotFound, "not found")
		return
	}
	if len(handler) == 0 {
		handler, _ = utils.GetCateHandler(p.Cate)
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg": "OK",
		"result": map[string]interface{}{
			"timeToRun": timeToRun,
			"params": params,
			"handler": handler,
		},
	})
}
