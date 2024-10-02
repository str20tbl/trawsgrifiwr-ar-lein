package controllers

import (
	"app/app/appJobs"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"log"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	*revel.Controller
}

func (c *App) Index() revel.Result {
	return c.Render()
}

func (c *App) Docs() revel.Result {
	return c.Render()
}

func (c *App) MergeSegment(uuid string, idA, idB int) revel.Result {
	data := fetchJSON(uuid)
	data.Transcripts[idA].End = data.Transcripts[idB].End
	data.Transcripts[idA].Text += " " + data.Transcripts[idB].Text
	data.Transcripts[idA].Words = append(data.Transcripts[idA].Words, data.Transcripts[idB].Words...)
	data.Transcripts = append(data.Transcripts[:idB], data.Transcripts[idB+1:]...)
	for i, _ := range data.Transcripts {
		data.Transcripts[i].ID = i
	}
	data.WriteJSON()
	return c.Redirect(c.Request.Referer())
}

func fetchJSON(uuid string) (data appJobs.ProcessFiles) {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uuid))
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	return
}

func (c *App) ExportSRT(uuid string) revel.Result {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uuid))
	var data appJobs.ProcessFiles
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	srtData := ""
	for _, el := range data.Transcripts {
		start := time.Unix(0, 0).UTC().Add(time.Duration(el.Start * float64(time.Second))).Format("T15:04:05.999Z")
		end := time.Unix(0, 0).UTC().Add(time.Duration(el.End * float64(time.Second))).Format("T15:04:05.999Z")
		srtData += fmt.Sprintf("%d\n%s --> %s\n%s\n\n", el.ID, start, end, el.Text)
	}
	srtFilename := fmt.Sprintf("/data/recordings/%s.srt", uuid)
	if err := os.WriteFile(srtFilename, []byte(srtData), 0666); err != nil {
		log.Fatal(err)
	}
	return c.RenderFileName(srtFilename, revel.Attachment)
}

func (c *App) UpdateJSON() revel.Result {
	var jsonData struct {
		UUID string             `json:"uuid"`
		Data appJobs.Transcript `json:"data"`
	}
	c.Params.BindJSON(&jsonData)
	revel.AppLog.Info(jsonData.UUID)
	uid := fmt.Sprintf("%s", jsonData.UUID)
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uid))
	var originalJSON appJobs.ProcessFiles
	err := json.Unmarshal(plan, &originalJSON)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	originalJSON.Transcripts = jsonData.Data
	originalJSON.WriteJSON()
	return c.RenderJSON(`{"success": "True"}`)
}

func (c *App) Editor(uuid string) revel.Result {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uuid))
	var data appJobs.ProcessFiles
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	return c.Render(data)
}

func (c *App) PlayAudio(uuid string) revel.Result {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.mp3", uuid))
	return c.RenderFile(audio, revel.Attachment)
}

func (c *App) Correct(uuid string) revel.Result {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uuid))
	var data appJobs.ProcessFiles
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	return c.Render(data)
}

const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
)

func (c *App) Upload(file []byte) revel.Result {
	// Validation rules.
	c.Validation.Required(file)
	c.Validation.MinSize(file, 2*KB).
		Message(`Minimum file size 2KB`)
	c.Validation.MaxSize(file, 200*MB).
		Message(`Max file size 200MB`)

	// Handle errors.
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect((*App).Index)
	}

	fileUUID := uuid.New()
	fileExt := filepath.Ext(c.Params.Files["file"][0].Filename)
	filename := fmt.Sprintf("/data/recordings/%s%s", fileUUID.String(), fileExt)
	err := os.WriteFile(filename, file, 0644)
	if err != nil {
		c.Validation.Error("Uwchlwytho ffeil wedi methu || File upload failed.")
		c.FlashParams()
		return c.Redirect((*App).Index)
	}
	job := appJobs.ProcessFiles{
		ContentType:      c.Params.Files["file"][0].Header.Get("Content-Type"),
		OriginalFilename: c.Params.Files["file"][0].Filename,
		Filename:         filename,
		UUID:             fileUUID.String(),
		Size:             float64(len(file)) / float64(KB),
		Step:             1,
		Status:           true,
	}
	job.WriteJSON()
	jobs.Now(job)

	return c.Redirect((*App).Correct, fileUUID.String())
}
