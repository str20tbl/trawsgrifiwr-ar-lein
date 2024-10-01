package controllers

import (
	"app/app/appJobs"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"os"
	"path/filepath"
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

func (c *App) PlayAudio(uuid string) {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.json", uuid))
	c.RenderFile(audio, revel.Attachment)
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
