package appJobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/revel/revel"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type ProcessFiles struct {
	ContentType      string
	OriginalFilename string
	Filename         string
	UUID             string
	Size             float64
	Step             int
	Status           bool
	Transcripts      Transcript
}

func (p ProcessFiles) Run() {
	//revel.AppLog.Info(p.Filename)
	mp3Filename := fmt.Sprintf("/data/recordings/%s.mp3", p.UUID)
	p.Status = runCommand(fmt.Sprintf("ffmpeg -i %s -vn -ar 44100 -ac 1 -b:a 192k %s", p.Filename, mp3Filename))
	p.Step += 1
	p.WriteJSON()
	if !p.Status {
		revel.AppLog.Error("Failed to convert file")
	} else {
		resp, err := transcribe(mp3Filename)
		if err != nil {
			revel.AppLog.Error("Failed to transcribe file", err)
		}
		//revel.AppLog.Infof("%+s", resp["id"])
		transcriptID := fmt.Sprintf("%s", resp["id"])
		for {
			time.Sleep(30 * time.Second)
			gotTranscript, err := getStatus(transcriptID)
			//revel.AppLog.Infof("Checking transcript %s ready: %t", transcriptID, gotTranscript)
			if err != nil {
				revel.AppLog.Error("Failed to transcribe file", err)
			}
			if gotTranscript {
				break
			}
		}
		p.Transcripts, err = getVAD(transcriptID)
		if err != nil {
			revel.AppLog.Error("Failed to transcribe file", err)
		}
		p.Step += 1
		p.WriteJSON()
		time.Sleep(5 * time.Second)
		p.Step += 1
		p.WriteJSON()
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
		revel.AppLog.Error(cmd)
		revel.AppLog.Error(string(cmdOut), err)
		return false
	}
	//revel.AppLog.Info(string(cmdOut))
	return true
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

	req, err := http.NewRequest("POST", "https://api-dev.techiaith.cymru/speech-to-text/v1/transcribe_long_form/?api_key=11111", &b)
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
	requestString := fmt.Sprintf("https://api-dev.techiaith.cymru/speech-to-text/v1/get_status/?stt_id=%s&api_key=11111", uuid)
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

type Transcript []struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
	Words []struct {
		Text      string    `json:"text"`
		Timestamp []float64 `json:"timestamp"`
	} `json:"words"`
}

func getVAD(uuid string) (Transcript, error) {
	requestString := fmt.Sprintf("https://api-dev.techiaith.cymru/speech-to-text/v1/get_vad_json/?stt_id=%s&api_key=11111", uuid)
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return nil, err
	}

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

	var response Transcript
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	//revel.AppLog.Infof("%+s", response)

	return response, nil
}
