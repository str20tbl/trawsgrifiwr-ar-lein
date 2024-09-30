package controllers

import (
	"fmt"
	"github.com/google/uuid"
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

func (c *App) Correct() revel.Result {
	return c.Render()
}

const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
)

type FileInfo struct {
	ContentType string
	Filename    string
	UUID        string
	Size        float64
	Status      string `json:",omitempty"`
}

func (c *App) Upload(file []byte) revel.Result {
	// Validation rules.
	c.Validation.Required(file)
	c.Validation.MinSize(file, 2*KB).
		Message("Minimum a file size of 2KB expected")
	c.Validation.MaxSize(file, 200*MB).
		Message("File cannot be larger than 200MB")

	// Handle errors.
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect((*App).Index)
	}

	fileUUID := uuid.New()
	fileExt := filepath.Ext(c.Params.Files["file"][0].Filename)
	err := os.WriteFile(fmt.Sprintf("/data/recordings/%s.%s", fileUUID.String(), fileExt), file, 0644)
	if err != nil {
		c.Validation.Error("Uwchlwytho ffeil wedi methu || File upload failed.")
		c.FlashParams()
		return c.Redirect((*App).Index)
	}

	return c.RenderJSON(FileInfo{
		ContentType: c.Params.Files["file"][0].Header.Get("Content-Type"),
		Filename:    c.Params.Files["file"][0].Filename,
		UUID:        fileUUID.String(),
		Size:        float64(len(file)) / float64(KB),
		Status:      "Successfully uploaded!",
	})
}
