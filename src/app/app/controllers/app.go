package controllers

import (
	"app/app/appJobs"
	"app/app/models"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"github.com/rogpeppe/go-internal/lockedfile"
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

func (c *App) RevertJSON(uuid string) revel.Result {
	backup := fetchBackup(uuid)
	backup.WriteJSON(false)
	return c.Redirect(c.Request.Referer())
}

func (c *App) AddSegment() revel.Result {
	var jsonData struct {
		UUID string `json:"uuid"`
		ID   int    `json:"id"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Error("Unable to bind JSON", err)
	}
	data := fetchJSON(jsonData.UUID)
	data.Transcripts.NewSegment(jsonData.ID, data.Duration)
	data.WriteJSON(false)
	output := renderTemplate(data)
	return c.RenderJSON(output.String())
}

func (c *App) MergeSegment() revel.Result {
	var jsonData struct {
		UUID string `json:"uuid"`
		IDa  int    `json:"idA"`
		IDb  int    `json:"idB"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Error("Unable to bind JSON", err)
	}
	revel.AppLog.Infof("%+v", jsonData)
	data := fetchJSON(jsonData.UUID)
	data.Transcripts.MergeSegments(jsonData.IDa, jsonData.IDb)
	data.WriteJSON(false)
	output := renderTemplate(data)
	return c.RenderJSON(output.String())
}

func (c *App) GetData() revel.Result {
	var jsonData struct {
		UUID string `json:"uuid"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Error("Unable to bind JSON", err)
	}
	data := fetchJSON(jsonData.UUID)
	return c.RenderJSON(data.AsJSON())
}

func fetchBackup(uid string) (data appJobs.ProcessFiles) {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.backup.json", uid))
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Errorf("Unable to open JSON %s, %v", uid, err)
	}
	return
}

func fetchJSON(uid string) (data appJobs.ProcessFiles) {
	plan, _ := lockedfile.Read(fmt.Sprintf("/data/recordings/%s.json", uid))
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Errorf("Unable to open JSON %s, %v", uid, err)
	}
	return
}

func (c *App) ExportSRT(uuid string) revel.Result {
	data := fetchJSON(uuid)
	srtFilename := data.Transcripts.ExportSRT(uuid)
	return c.RenderFileName(srtFilename, revel.Attachment)
}

// UpdateJSON save a copy of the current JSON file to disk
func (c *App) UpdateJSON() revel.Result {
	var jsonData struct {
		UUID string            `json:"uuid"`
		Data models.Transcript `json:"data"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Errorf("Unable to bind JSON :: %v", err)
	}
	originalJSON := fetchJSON(jsonData.UUID)
	originalJSON.Transcripts = jsonData.Data
	originalJSON.Updated = time.Now().Format("2006-01-02 15:04")
	if originalJSON.Started == "0001-01-01T00:00:00Z" || originalJSON.Started == "" {
		originalJSON.Started = time.Now().Format("2006-01-02 15:04")
	}
	originalJSON.WriteJSON(false)
	output := renderTemplate(originalJSON)
	return c.RenderJSON(output.String())
}

// Editor displays the transcription editor view to the user
func (c *App) Editor(uuid string, segmentID int) revel.Result {
	data := fetchJSON(uuid)
	if data.UUID == "" {
		return c.Redirect("/")
	}
	return c.Render(data, segmentID, UUIDBlackList)
}

// PlayAudio from a file on disk, serve the audio file
func (c *App) PlayAudio(uuid string) revel.Result {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.wav", uuid))
	return c.RenderFile(audio, revel.Attachment)
}

// Transcribe show the transcription progress to the user
func (c *App) Transcribe(uuid string) revel.Result {
	data := fetchJSON(uuid)
	return c.Render(data)
}

// DeleteRecord delete all data associated with the given UUID
func (c *App) DeleteRecord(UUID string) revel.Result {
	base := "/data/recordings/%s.%s"
	err := os.Remove(fmt.Sprintf(base, UUID, "json"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, UUID, "mp3"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, UUID, "wav"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, UUID, "backup.json"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	return c.Redirect("/")
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
	c.Validation.MaxSize(file, 746*MB).
		Message(`Max file size 746MB`)

	// Handle errors.
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect((*App).Index)
	}

	fileUUID := uuid.New()
	fileUUIDt := uuid.New()
	fileExt := filepath.Ext(c.Params.Files["file"][0].Filename)
	filename := fmt.Sprintf("/data/recordings/%s%s", fileUUID.String(), fileExt)
	filenamet := fmt.Sprintf("/data/recordings/%s%s", fileUUIDt.String(), fileExt)
	err := os.WriteFile(filename, file, 0644)
	if err != nil {
		c.Validation.Error("Uwchlwytho ffeil wedi methu || File upload failed.")
		c.FlashParams()
		return c.Redirect((*App).Index)
	}
	err = os.WriteFile(filenamet, file, 0644)
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
		UUIDTemp:         filenamet,
		Size:             float64(len(file)) / float64(KB),
		Step:             1,
		Status:           true,
	}
	job.WriteJSON(false)
	jobs.Now(job)

	return c.Redirect((*App).Transcribe, fileUUID.String())
}
