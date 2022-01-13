// global conf
// ENV:
//   CONF_FILE      --- configuration file name
//   TZ             --- Time Zone name, e.g. "Asia/Shanghai"
//
// JSON of <CONF_FILE>:
// {
//      "listen-host": "",
//      "listen-port": 7080,
//      "cb-retry-times": 3,
//      "saving-home": "home path to save/restore"
// }
//
// Rosbit Xu
package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"path"
)

type DelayTaskConf struct {
	ListenHost    string `json:"listen-host"`
	ListenPort    int    `json:"listen-port"`
	CbRetryTimes  int    `json:"cb-retry-times"`
	SavingHome    string `json:"saving-home"`
}

var (
	ServiceConf DelayTaskConf
	Loc = time.FixedZone("UTC+8", 8*60*60)
)

func getEnv(name string, result *string, must bool) error {
	s := os.Getenv(name)
	if s == "" {
		if must {
			return fmt.Errorf("env \"%s\" not set", name)
		}
	}
	*result = s
	return nil
}

func CheckGlobalConf() error {
	var p string
	getEnv("TZ", &p, false)
	if p != "" {
		if loc, err := time.LoadLocation(p); err == nil {
			Loc = loc
		}
	}

	var confFile string
	if err := getEnv("CONF_FILE", &confFile, true); err != nil {
		return err
	}
	fp, err := os.Open(confFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	dec := json.NewDecoder(fp)
	if err = dec.Decode(&ServiceConf); err != nil {
		return err
	}

	if err = checkMust(confFile); err != nil {
		return err
	}

	return nil
}

func checkMust(confFile string) error {
	confRoot := path.Dir(confFile)

	if ServiceConf.ListenPort <= 0 {
		return fmt.Errorf("listening port expected in conf")
	}

	if ServiceConf.CbRetryTimes <= 0 || ServiceConf.CbRetryTimes > 5 {
		ServiceConf.CbRetryTimes = 3
	}

	if len(ServiceConf.SavingHome) == 0 {
		return fmt.Errorf("saving-home expected in conf")
	}
	ServiceConf.SavingHome = toAbsPath(confRoot, ServiceConf.SavingHome)
	if err := checkDir(ServiceConf.SavingHome, "saving-home"); err != nil {
		return err
	}

	return nil
}

func DumpConf() {
	fmt.Printf("conf: %#v\n", ServiceConf)
	fmt.Printf("TZ time location: %v\n", Loc)
}

func checkDir(path, prompt string) error {
	if fi, err := os.Stat(path); err != nil {
		return err
	} else if !fi.IsDir() {
		return fmt.Errorf("%s %s is not a directory", prompt, path)
	}
	return nil
}

func toAbsPath(absRoot, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}
	return path.Join(absRoot, filePath)
}
