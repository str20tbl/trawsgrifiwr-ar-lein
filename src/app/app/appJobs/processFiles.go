package appJobs

import (
	"app/app/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/revel"
	"github.com/rogpeppe/go-internal/lockedfile"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ProcessFiles struct {
	ContentType      string
	OriginalFilename string
	Filename         string
	UUID             string
	UUIDTemp         string
	Size             float64
	Step             int
	Status           bool
	Duration         float64
	Started          string
	Updated          string
	Transcripts      models.Transcript
	TranscriptID     string
}

func (p ProcessFiles) Run() {
	revel.AppLog.Info(p.Filename)
	p.WriteJSON(true)
	var err error
	mp3Filename := fmt.Sprintf("/data/recordings/%s.mp3", p.UUID)
	wavFilename := fmt.Sprintf("/data/recordings/%s.wav", p.UUID)
	_, _ = runCommand(fmt.Sprintf("ffmpeg -i %s -acodec pcm_s16le -ac 1 -ar 16000 %s", p.UUIDTemp, wavFilename))
	if !strings.HasSuffix(p.UUIDTemp, "mp3") {
		p.Status, _ = runCommand(fmt.Sprintf("ffmpeg -i %s %s", p.UUIDTemp, mp3Filename))
	} else {
		p.Status, _ = runCommand(fmt.Sprintf("cp %s %s", p.UUIDTemp, mp3Filename))
	}
	_, duration := runCommand(fmt.Sprintf("ffprobe -i %s -show_entries format=duration -v quiet -of csv='p=0'", p.Filename))
	p.Duration, err = strconv.ParseFloat(strings.Trim(duration, "\n"), 64)
	if err != nil {
		revel.AppLog.Error("Failed to convert file", err)
		p.Status = false
		return
	}
	p.Step += 1
	p.WriteJSON(true)
	if !p.Status {
		revel.AppLog.Error("Failed to convert file")
		return
	}
	resp, err := transcribe(mp3Filename)
	if err != nil {
		revel.AppLog.Error("Failed to convert file")
		p.Status = false
		return
	}
	var tID interface{}
	if tID, p.Status = resp["id"]; !p.Status {
		revel.AppLog.Error("Failed to convert file")
		return
	}
	p.TranscriptID = fmt.Sprintf("%s", tID)
	revel.AppLog.Infof("%+v", tID)
	_, err = uuid.Parse(p.TranscriptID)
	if err != nil {
		revel.AppLog.Error("Failed to convert file")
		p.Status = false
		return
	}
	p.WriteJSON(true)
	for {
		time.Sleep(30 * time.Second)
		gotTranscript, err := getStatus(p.TranscriptID)
		if err != nil {
			revel.AppLog.Errorf("Failed to transcribe file %+v", err)
		}
		if gotTranscript {
			break
		}
	}
	p.Transcripts, err = getVAD(p.TranscriptID)
	if err != nil {
		revel.AppLog.Error("Failed to transcribe file", err)
		p.Status = false
		return
	}
	for i := range p.Transcripts.Segments {
		p.Transcripts.Segments[i].ID = i
	}
	p.Step += 1
	p.WriteJSON(false)
	time.Sleep(5 * time.Second)
	p.Step += 1
	p.WriteJSON(false)
	p.SaveBackup()
}

func (p ProcessFiles) SaveBackup() {
	filepath := fmt.Sprintf("/data/recordings/%s.backup.json", p.UUID)
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

func (p ProcessFiles) AsJSON() string {
	asJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		revel.AppLog.Error("Unable to marshal JSON", err)
	}
	return string(asJSON)
}

func (p ProcessFiles) WriteJSON(overwrite bool) {
	if len(p.Transcripts.Segments) > 0 || overwrite {
		filepath := fmt.Sprintf("/data/recordings/%s.json", p.UUID)
		f, err := lockedfile.Create(filepath)
		if err != nil {
			revel.AppLog.Error("Unable to create file to save JSON", err)
		}
		if err == nil {
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
	}
}

func runCommand(cmd string) (bool, string) {
	cmdOut, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		revel.AppLog.Error(cmd)
		revel.AppLog.Error(string(cmdOut), err)
		return false, ""
	}
	revel.AppLog.Info(string(cmdOut))
	return true, string(cmdOut)
}

func transcribe(filename string) (map[string]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreateFormFile("soundfile", filename)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", "https://api-dev.techiaith.cymru/speech-to-text/v2/transcribe_long_form/?api_key=11111", &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func getStatus(uuid string) (bool, error) {
	requestString := fmt.Sprintf("https://api-dev.techiaith.cymru/speech-to-text/v2/get_status/?stt_id=%s&api_key=11111", uuid)
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return false, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response map[string]interface{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}
	status := fmt.Sprintf("%+s", response["status"])
	return status == "SUCCESS", nil
}

func getVAD(uuid string) (t models.Transcript, err error) {
	requestString := fmt.Sprintf("https://api-dev.techiaith.cymru/speech-to-text/v2/get_json/?stt_id=%s&api_key=11111", uuid)
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &t); err != nil {
		return
	}
	for i, _ := range t.Segments {
		t.Segments[i].ID = i
	}
	return
}
