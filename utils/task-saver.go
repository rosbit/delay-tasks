package utils

import (
	"delay-tasks/conf"
	"encoding/json"
	"os"
	"fmt"
	"path"
)

const (
	handler_file    = "delay-task-handlers.json"
	delay_task_file = "delay-tasks.json"
)

func restoreData() error {
	if err := restoreHandlers(); err != nil {
		return err
	}
	if err := restoreDelayTask(); err != nil {
		return err
	}
	return nil
}

func saveHandlers() error {
	handlerFile := fmt.Sprintf("%s/%s", conf.ServiceConf.SavingHome, handler_file)
	return saveFile(handlerFile, taskHandlers)
}

func restoreHandlers() error {
	handlerFile := fmt.Sprintf("%s/%s", conf.ServiceConf.SavingHome, handler_file)
	var handlers map[string]string
	if err := restoreFile(handlerFile, &handlers); err != nil {
		return err
	}
	if len(handlers) > 0 {
		taskHandlers = handlers
	}
	return nil
}

func saveDelayTasks() error {
	delayTaskFile := fmt.Sprintf("%s/%s", conf.ServiceConf.SavingHome, delay_task_file)
	return saveFile(delayTaskFile, delayTasks)
}

func restoreDelayTask() error {
	delayTaskFile := fmt.Sprintf("%s/%s", conf.ServiceConf.SavingHome, delay_task_file)
	var tasks []*cateTasksT
	if err := restoreFile(delayTaskFile, &tasks); err != nil {
		return err
	}
	if len(tasks) > 0 {
		delayTasks = tasks
	}
	return nil
}

func saveFile(dataFileName string, data interface{}) error {
	dataDir := path.Dir(dataFileName)
	dataFile := path.Base(dataFileName)
	fTmp, err := os.CreateTemp(dataDir, fmt.Sprintf("%s-******.tmp", dataFile))
	if err != nil {
		return err
	}
	fTmp.Chmod(0644)
	tmpFile := fTmp.Name()
	enc := json.NewEncoder(fTmp)
	enc.Encode(data)
	fTmp.Close()

	return os.Rename(tmpFile, dataFileName)
}

func restoreFile(dataFileName string, data interface{}) error {
	fp, err := os.Open(dataFileName)
	if err != nil {
		return err
	}
	defer fp.Close()

	return json.NewDecoder(fp).Decode(data)
}
