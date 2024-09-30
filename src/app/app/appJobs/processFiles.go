package appJobs

import (
	"fmt"
	"github.com/revel/revel"
	"os/exec"
)

type ProcessFiles struct {
	Filename string
}

func (p ProcessFiles) Run() {
	revel.AppLog.Info(p.Filename)
	success := runCommand(fmt.Sprintf("ffmpeg -i input.wav -vn -ar 44100 -ac 2 -b:a 192k %s", p.Filename))
	if !success {
		revel.AppLog.Info("Failed to convert file")
	}
}

func runCommand(cmd string) bool {
	cmdOut, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		revel.AppLog.Info(cmd)
		revel.AppLog.Infof(string(cmdOut))
		return false
	}
	revel.AppLog.Info(string(cmdOut))
	return true
}
