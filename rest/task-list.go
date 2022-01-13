package rest

import (
	"delay-tasks/utils"
	"net/http"
)

// GET /tasks
func ListTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	utils.ListTask(w)
}
