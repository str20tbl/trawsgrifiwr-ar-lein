package appJobs

import (
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"os"
	"os/exec"
)

type ProcessFiles struct {
	ContentType      string
	OriginalFilename string
	Filename         string
	UUID             string
	Size             float64
	Step             int
	Status           bool
}

func (p ProcessFiles) Run() {
	revel.AppLog.Info(p.Filename)
	p.Status = runCommand(fmt.Sprintf("ffmpeg -i %s -vn -ar 44100 -ac 2 -b:a 192k %s", p.Filename, fmt.Sprintf("/data/recordings/%s.mp3", p.UUID)))
	p.Step += 1
	p.WriteJSON()
	if !p.Status {
		revel.AppLog.Info("Failed to convert file")
	} else {

	}
}

func (p ProcessFiles) WriteJSON() {
	filepath := fmt.Sprintf("/data/recordings/%s.json", p.UUID)
	f, err := os.Create(filepath)
	if err != nil {
		revel.AppLog.Error("Unable to create file to save JSON", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			revel.AppLog.Error("Unable to close JSON file", err)
		}
	}()
	asJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		revel.AppLog.Error("Unable to marshal JSON", err)
	}
	_, err = f.Write(asJSON)
	if err != nil {
		revel.AppLog.Error("Unable to save JSON", err)
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
