package controllers

import (
	"app/app/appJobs"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"html/template"
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

func (c *App) RevertJSON(uuid string) revel.Result {
	backup := fetchBackup(uuid)
	backup.WriteJSON()
	return c.Redirect(c.Request.Referer())
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
	data.Transcripts.Segments[jsonData.IDa].End = data.Transcripts.Segments[jsonData.IDb].End
	data.Transcripts.Segments[jsonData.IDa].Text += " " + data.Transcripts.Segments[jsonData.IDb].Text
	data.Transcripts.Segments[jsonData.IDa].Words = append(data.Transcripts.Segments[jsonData.IDa].Words, data.Transcripts.Segments[jsonData.IDb].Words...)
	data.Transcripts.Segments = append(data.Transcripts.Segments[:jsonData.IDb], data.Transcripts.Segments[jsonData.IDb+1:]...)
	for i, _ := range data.Transcripts.Segments {
		data.Transcripts.Segments[i].ID = i
	}
	data.WriteJSON()
	page := `{{range .data.Transcripts.Segments}}<div class="card m-3 pt-3 row hide transcript" id="{{.ID}}_tran">
	<table class="table">
		<tr>
			<td>ID: {{.ID}}</td>
			<td>Start: {{.Start}}</td>
			<td>End: {{.End}}</td>
			{{if gt .ID 0}}
			<td>
				<button class="btn btn-sm btn-outline-primary" onclick="mergeSegments({{add .ID -1}}, {{.ID}})">
					Cyfuno gyda
					ID: {{add .ID -1}}
				</button>
			</td>
			{{end}}
		</tr>
	</table>
	<div class="col-sm-12 pb-3">
		<textarea class="form-control" id="staticEmail2" rows="5" onfocusout="saveFile()">{{.Text}}</textarea>
	</div>
</div>
{{end}}`
	tmpl, err := template.New("page").Funcs(template.FuncMap{
		"add": func(a, b int) string {
			return fmt.Sprintf("%d", a+b)
		},
	}).Parse(page)
	if err != nil {
		revel.AppLog.Error("Parse", err)
	}
	var output bytes.Buffer
	err = tmpl.Execute(&output, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		revel.AppLog.Error("Execute", err)
	}
	return c.RenderJSON(output.String())
}

func fetchBackup(uuid string) (data appJobs.ProcessFiles) {
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.backup.json", uuid))
	err := json.Unmarshal(plan, &data)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	return
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
	data := fetchJSON(uuid)
	srtData := ""
	for _, el := range data.Transcripts.Segments {
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
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Error("Unable to bind JSON", err)
	}
	revel.AppLog.Info(jsonData.UUID)
	uid := fmt.Sprintf("%s", jsonData.UUID)
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uid))
	var originalJSON appJobs.ProcessFiles
	err = json.Unmarshal(plan, &originalJSON)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	originalJSON.Transcripts = jsonData.Data
	originalJSON.Updated = time.Now()
	if originalJSON.Started.IsZero() {
		originalJSON.Started = time.Now()
	}
	originalJSON.WriteJSON()
	return c.RenderJSON(`{"success": "True"}`)
}

func (c *App) Editor(uuid string) revel.Result {
	data := fetchJSON(uuid)
	return c.Render(data)
}

func (c *App) PlayAudio(uuid string) revel.Result {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.mp3", uuid))
	return c.RenderFile(audio, revel.Attachment)
}

func (c *App) Correct(uuid string) revel.Result {
	data := fetchJSON(uuid)
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
	c.Validation.MaxSize(file, 746*MB).
		Message(`Max file size 746MB`)

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
