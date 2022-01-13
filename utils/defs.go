package utils

import (
	"time"
)

const (
	OP_ADD = iota
	OP_DEL
	OP_REGISTER_HANDLER
)

type (
	taskT struct {
		TimeToRun int64
		Params interface{}
		Handler string
	}

	cateTasksT struct {
		LastExecTime int64 // last executing timestamp
		Tasks map[string]map[uint64]*taskT  // cate name -> key -> Task
	}

	addTaskT struct {
		execTime *time.Time
		cate string
		key uint64
		params interface{}
		handler string
	}
	delTaskT struct {
		execTime *time.Time
		cate string
		key uint64
		exec bool
	}
	registerCateT struct {
		cate string
		handler string
	}

	opT struct {
		opType int
		params interface{}
	}
)

