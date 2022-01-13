package utils

import (
	"github.com/rosbit/gnet"
	"delay-tasks/conf"
	"encoding/json"
	"time"
	"sync"
	"io"
	"fmt"
)

const (
	SLOT_TIME_SPAN     = 60    // seconds of slot time span
	MAX_ONCE_LOOP_TIME = 3600  // time in seconds to finish a loop
	SLOT_COUNT = MAX_ONCE_LOOP_TIME / SLOT_TIME_SPAN
)

var (
	taskHandlers = map[string]string{} // cate name -> handler url
	delayTasks []*cateTasksT           // one slot per minute
	opChan chan *opT
	taskMutex = &sync.Mutex{}
	handlerMutex = &sync.Mutex{}

	lastCheckSlot = -2
)

func RegisterCateHandler(cate string, handler string) {
	opChan <- &opT{
		opType: OP_REGISTER_HANDLER,
		params: &registerCateT{
			cate: cate,
			handler: handler,
		},
	}
}

func GetCateHandler(cate string) (handler string, ok bool) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	handler, ok = taskHandlers[cate]
	return
}

func AddTask(execTime time.Time, cate string, key uint64, params interface{}, handler string) {
	opChan <- &opT{
		opType: OP_ADD,
		params: &addTaskT{
			execTime: &execTime,
			cate: cate,
			key: key,
			params: params,
			handler: handler,
		},
	}
}

func DelTask(execTime time.Time, cate string, key uint64, exec bool) {
	opChan <- &opT{
		opType: OP_DEL,
		params: &delTaskT{
			execTime: &execTime,
			cate: cate,
			key: key,
			exec: exec,
		},
	}
}

func timeSlot(execTime *time.Time) int {
	return execTime.Minute()
}

func registerCateHandler(p *registerCateT) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	taskHandlers[p.cate] = p.handler

	saveHandlers()
}

func addTask(p *addTaskT) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	taskSlot := timeSlot(p.execTime)
	// fmt.Printf("taskSlot in addTask: %d\n", taskSlot)
	slotTasks := delayTasks[taskSlot]
	if slotTasks == nil {
		slotTasks = &cateTasksT{
			Tasks: map[string]map[uint64]*taskT{},
		}
		delayTasks[taskSlot] = slotTasks
	}

	lastCheckSlot -= 1
	slotTasks.LastExecTime =  time.Now().Unix() - MAX_ONCE_LOOP_TIME // make chance to run task if p.execTime is near current time

	cateTasks, ok := slotTasks.Tasks[p.cate]
	if !ok || cateTasks == nil {
		cateTasks = map[uint64]*taskT{}
		slotTasks.Tasks[p.cate] = cateTasks
	}
	cateTasks[p.key] = &taskT{
		TimeToRun: p.execTime.Unix() / SLOT_TIME_SPAN * SLOT_TIME_SPAN, // truncate seconds
		Params: p.params,
		Handler: p.handler,
	}

	saveDelayTasks()
}

func delTask(p *delTaskT) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	taskSlot := timeSlot(p.execTime)
	slotTasks := delayTasks[taskSlot]
	if slotTasks == nil || len(slotTasks.Tasks) == 0 {
		return
	}

	cateTasks, ok := slotTasks.Tasks[p.cate]
	if !ok || len(cateTasks) == 0 {
		return
	}
	task, ok := cateTasks[p.key]
	if !ok {
		return
	}

	if p.exec {
		// exec task before being removed.
		handler := task.Handler
		if len(handler) == 0 {
			handler, _ = GetCateHandler(p.cate)
		}
		if len(handler) > 0 {
			go runTask(handler, p.cate, p.key, task.Params, true)
		}
	}
	delete(cateTasks, p.key)

	if len(cateTasks) == 0 {
		delete(slotTasks.Tasks, p.cate)
		if len(slotTasks.Tasks) == 0 {
			delayTasks[taskSlot] = nil
		}
	}

	saveDelayTasks()
}

func taskManagerThread() {
	for op := range opChan {
		switch op.opType {
		case OP_ADD:
			p := op.params.(*addTaskT)
			addTask(p)
		case OP_REGISTER_HANDLER:
			p := op.params.(*registerCateT)
			registerCateHandler(p)
		case OP_DEL:
			p := op.params.(*delTaskT)
			delTask(p)
		default:
		}
	}
}

func timeLoopThread() {
	ticker := time.NewTicker(time.Minute)
	for now := range ticker.C {
		taskSlot := timeSlot(&now)
		timestamp := now.Unix()

		// fmt.Printf("taskSlot/now in timeLoop: %d/%d\n", taskSlot, timestamp)
		if taskSlot - lastCheckSlot > 1 {
			// check the last slot to make sure NOT miss running the task.
			lastSlot := func(taskSlot int)int {
				taskSlot -= 1
				if taskSlot < 0 {
					return SLOT_COUNT-1
				}
				return taskSlot
			}(taskSlot)

			checkAndRunSlot(lastSlot, timestamp, true)
			// fmt.Printf("lastSlot #%d check\n", lastSlot)
		}

		// let other operation have chance to run
		checkAndRunSlot(taskSlot, timestamp, false)
		// fmt.Printf("taskSlot #%d check\n", taskSlot)

		lastCheckSlot = func(taskSlot int)int {
			if taskSlot == SLOT_COUNT-1 {
				return -1
			}
			return taskSlot
		}(taskSlot)
	}
}

func checkAndRunSlot(taskSlot int, timestamp int64, checkFirst bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	// fmt.Printf("in checkAndRunSlot(%d, %d, %v)\n", taskSlot, timestamp, checkFirst)

	slotTasks := delayTasks[taskSlot]
	if slotTasks == nil {
		return
	}
	if len(slotTasks.Tasks) == 0 {
		delayTasks[taskSlot] = nil
		saveDelayTasks()
		return
	}

	// fmt.Printf(" slot #%d execTime %d\n", taskSlot, slotTasks.LastExecTime)
	if checkFirst && timestamp - slotTasks.LastExecTime < MAX_ONCE_LOOP_TIME {
		// the task is executed, just return
		return
	}

	slotTasks.LastExecTime = timestamp // set task flag of being executed
	deleteCates := make([]string, len(slotTasks.Tasks))
	deleteCateCount := 0
	for cate, cateTasks := range slotTasks.Tasks {
		if len(cateTasks) == 0 {
			deleteCates[deleteCateCount] = cate
			deleteCateCount += 1
			continue
		}
		cateHandler, _ := GetCateHandler(cate)

		deleteKeys := make([]uint64, len(cateTasks))
		deleteKeyCount := 0
		for key, task := range cateTasks {
			// fmt.Printf(" timestamp: %d, task.timestamp: %d\n", timestamp, task.TimeToRun)
			if task.TimeToRun > timestamp {
				continue
			}

			handler := task.Handler
			if len(handler) == 0 {
				handler = cateHandler
			}

			if len(handler) > 0 {
				go runTask(handler, cate, key, task.Params)
			}
			deleteKeys[deleteKeyCount] = key
			deleteKeyCount += 1
		}

		if deleteKeyCount >= len(cateTasks) {
			deleteCates[deleteCateCount] = cate
			deleteCateCount += 1
		} else {
			for i:=0; i<deleteCateCount; i++ {
				delete(cateTasks, deleteKeys[i])
			}
		}
	}

	if deleteCateCount >= len(slotTasks.Tasks) {
		delayTasks[taskSlot] = nil
	} else {
		for i:=0; i<deleteCateCount; i++ {
			delete(slotTasks.Tasks, deleteCates[i])
		}
	}

	saveDelayTasks()
}

func runTask(handler, cate string, key uint64, params interface{}, inAdvance ...bool) {
	handlerParams := map[string]interface{}{
		"cate": cate,
		"key": key,
		"params": params,
		"inAdvance": func()bool{if len(inAdvance) > 0 {return inAdvance[0]}; return false}(),
	}

	for i:=0; i<conf.ServiceConf.CbRetryTimes; i++ {
		status, content, _, err := gnet.JSON(handler, gnet.Params(handlerParams))

		if err != nil {
			fmt.Printf("failed to callback %s(cate:%s, key:%d) #%d time(s): %v\n", handler, cate, key, i, err)
			continue
		}
		fmt.Printf("callback %s(cate:%s, key:%d) result: status %d, content: %s\n", handler, cate, key, status, content)
		break
	}
}

func StartLoop() {
	// fmt.Printf("SLOT_COUNT: %d\n", SLOT_COUNT)
	delayTasks = make([]*cateTasksT, SLOT_COUNT)
	opChan = make(chan *opT, 5)
	restoreData()
	go taskManagerThread()
	go timeLoopThread()
}

func toLocalTime(timestamp int64) string {
	return time.Unix(timestamp, 0).In(conf.Loc).Format(time.ANSIC)
}

func listTask(i int, slotTasks *cateTasksT, out io.Writer) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	fmt.Fprintf(out, "slot #%d, last check time: %s\n", i, toLocalTime(slotTasks.LastExecTime))
	enc := json.NewEncoder(out)

	found := false
	for cate, cateTasks := range slotTasks.Tasks {
		fmt.Fprintf(out, " +- cate: %s\n", cate)
		if handler, ok := taskHandlers[cate]; ok {
			fmt.Fprintf(out, "  +- [handler] %s\n", handler)
		}
		if len(cateTasks) == 0 {
			io.WriteString(out, "  +- [no tasks]\n")
			continue
		}
		found = true
		for key, task := range cateTasks {
			fmt.Fprintf(out, "  +- [task] key: %d, time to exec: %s, task handler: %s, params:", key, toLocalTime(task.TimeToRun), task.Handler)
			enc.Encode(task.Params)
		}
	}
	if !found {
		io.WriteString(out, " +- [no cate]\n")
	}
}

func ListTask(out io.Writer) {
	fmt.Fprintf(out, "SLOT_TIME_SPAN: %d\n", SLOT_TIME_SPAN)
	fmt.Fprintf(out, "lastCheckSlot: %d\n", lastCheckSlot)
	found := false
	for i, slotTasks := range delayTasks {
		if slotTasks == nil {
			continue
		}
		listTask(i, slotTasks, out)
		found = true
	}
	if !found {
		io.WriteString(out, "[no slots]\n")
	}
}

func findTaskInSlot(slotTasks *cateTasksT, cate string, key uint64) *taskT {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	cateTasks, ok := slotTasks.Tasks[cate]
	if !ok || len(cateTasks) == 0 {
		return nil
	}
	task, ok := cateTasks[key]
	if !ok {
		return nil
	}
	return task
}

func FindTask(cate string, key uint64, timestamp int64) (timeToRun int64, params interface{}, handler string, ok bool) {
	if timestamp <= 0 {
		// look for the first satisfied task
		for _, slotTasks := range delayTasks {
			if slotTasks == nil {
				continue
			}
			if task := findTaskInSlot(slotTasks, cate, key); task != nil {
				return task.TimeToRun, task.Params, task.Handler, true
			}
		}
		return
	}

	// look for task in the specified slots
	execTime := time.Unix(timestamp, 0)
	taskSlot := timeSlot(&execTime)
	slotTasks := delayTasks[taskSlot]
	if slotTasks == nil {
		return
	}
	if task := findTaskInSlot(slotTasks, cate, key); task != nil {
		return task.TimeToRun, task.Params, task.Handler, true
	}
	return
}

