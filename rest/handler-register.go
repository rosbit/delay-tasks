package rest

import (
	"github.com/rosbit/mgin"
	"delay-tasks/utils"
	"net/http"
	"strings"
)

// POST /handler/:cate
// BODY: {"handler": "url"}
func RegisterCateHandler(c *mgin.Context) {
	var params struct {
		Handler string
	}
	cate := c.Param("cate")
	if status, err := c.ReadJSON(&params); err != nil {
		c.Error(status, err.Error())
		return
	}
	if !strings.HasPrefix(params.Handler, "http") {
		c.Error(http.StatusBadRequest, "handler param expected")
		return
	}
	utils.RegisterCateHandler(cate, params.Handler)
	c.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg": "OK",
	})
}
