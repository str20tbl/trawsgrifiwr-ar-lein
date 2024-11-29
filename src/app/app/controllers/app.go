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

var tmlpFuncs = template.FuncMap{
	"add": func(a, b int) string {
		return fmt.Sprintf("%d", a+b)
	},
	"addTime": func(a, b float64) float64 {
		return a + b
	},
	"s2m": func(inSeconds float64) string {
		return time.Duration(inSeconds * float64(time.Second)).String()
	},
}

const page = `{{range .data.Transcripts.Segments}}
                        <div id="{{.ID}}_text" class="hide m-3 transcript">
                            <div class="card">
                                <div class="card-body">
                                    <div class="row">
                                        <div class="col-1">
                                            ID: {{.ID}}
                                        </div>
                                        <div class="col-9">
                                            {{.Text}}
                                        </div>
                                        <div class="col-2">
                                            {{if gt .ID 0}}
                                            <button class="btn btn-sm btn-outline-primary" onclick="mergeSegments({{add .ID -1}}, {{.ID}})">
                                                Cyfuno gyda
                                                ID: {{add .ID -1}}
                                            </button>
                                            {{else}}
                                            &nbsp
                                            {{end}}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="card m-3 pt-3 row hide transcript" id="{{.ID}}_tran">
                            <table class="table">
                                <tr>
                                    <td>
                                        <button class="btn btn-sm btn-outline-primary" onclick="restart_segment({{addTime .Start -0.25}})">
                                            <svg width="18px" height="18px" xmlns="http://www.w3.org/2000/svg" fill="currentColor" class="bi bi-play-fill" viewBox="0 0 16 16">
                                              <path d="m11.596 8.697-6.363 3.692c-.54.313-1.233-.066-1.233-.697V4.308c0-.63.692-1.01 1.233-.696l6.363 3.692a.802.802 0 0 1 0 1.393z"></path>
                                            </svg>
                                        </button>
                                        <button class="btn btn-sm btn-outline-primary" id="repeat-{{.ID}}" onclick="loop_segment('repeat-{{.ID}}', {{addTime .Start -0.25}}, {{addTime .End 0.25}})">
                                            <svg width="18px" height="18px" xmlns="http://www.w3.org/2000/svg" fill="currentColor" class="bi bi-recycle" viewBox="0 0 16 16">
                                              <path d="M9.302 1.256a1.5 1.5 0 0 0-2.604 0l-1.704 2.98a.5.5 0 0 0 .869.497l1.703-2.981a.5.5 0 0 1 .868 0l2.54 4.444-1.256-.337a.5.5 0 1 0-.26.966l2.415.647a.5.5 0 0 0 .613-.353l.647-2.415a.5.5 0 1 0-.966-.259l-.333 1.242-2.532-4.431zM2.973 7.773l-1.255.337a.5.5 0 1 1-.26-.966l2.416-.647a.5.5 0 0 1 .612.353l.647 2.415a.5.5 0 0 1-.966.259l-.333-1.242-2.545 4.454a.5.5 0 0 0 .434.748H5a.5.5 0 0 1 0 1H1.723A1.5 1.5 0 0 1 .421 12.24l2.552-4.467zm10.89 1.463a.5.5 0 1 0-.868.496l1.716 3.004a.5.5 0 0 1-.434.748h-5.57l.647-.646a.5.5 0 1 0-.708-.707l-1.5 1.5a.498.498 0 0 0 0 .707l1.5 1.5a.5.5 0 1 0 .708-.707l-.647-.647h5.57a1.5 1.5 0 0 0 1.302-2.244l-1.716-3.004z"></path>
                                            </svg>
                                        </button>
                                    </td>
                                    <td>ID: {{.ID}}</td>
                                    <td>
                                        Start: {{s2m .Start}}
                                    </td>
                                    <td>
                                        End: {{s2m .End}}
                                    </td>
                                    <td>
                                        Segment Newydd
                                        <br>
                                        <button class="btn btn-sm btn-outline-primary" onclick="addSegment({{add .ID -1}})">
                                            cyn
                                        </button>
                                        <button class="btn btn-sm btn-outline-primary" onclick="addSegment({{.ID}})">
                                            ar ôl
                                        </button>
                                    </td>
                                </tr>
                            </table>
                            <div class="col-sm-12 pb-3">
                                <textarea class="form-control" id="staticEmail2" rows="5" onfocusout="saveFile()">{{.Text}}</textarea>
                            </div>
                        </div>
                        {{end}}`

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

func (c *App) AddSegment() revel.Result {
	var jsonData struct {
		UUID string `json:"uuid"`
		ID   int    `json:"id"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Error("Unable to bind JSON", err)
	}
	revel.AppLog.Infof("%+v", jsonData)
	data := fetchJSON(jsonData.UUID)
	endTime := data.Duration
	if jsonData.ID < len(data.Transcripts.Segments)-1 {
		endTime = data.Transcripts.Segments[jsonData.ID+1].Start - 0.1
	}
	startTime := 0.0
	if jsonData.ID > 0 {
		startTime = data.Transcripts.Segments[jsonData.ID].End + 0.1
	}
	newSeg := appJobs.Segment{
		ID:    0,
		Start: startTime,
		End:   endTime,
		Text:  "",
	}
	if jsonData.ID >= len(data.Transcripts.Segments)-1 {
		data.Transcripts.Segments = append(data.Transcripts.Segments, newSeg)
	} else if jsonData.ID == -1 {
		data.Transcripts.Segments = append([]appJobs.Segment{newSeg}, data.Transcripts.Segments...)
	} else {
		tList := make([]appJobs.Segment, 0)
		for i, el := range data.Transcripts.Segments {
			tList = append(tList, el)
			if i == jsonData.ID {
				tList = append(tList, newSeg)
			}
		}
		data.Transcripts.Segments = tList
	}
	for i, _ := range data.Transcripts.Segments {
		data.Transcripts.Segments[i].ID = i
	}
	data.WriteJSON()
	tmpl, err := template.New("page").Funcs(tmlpFuncs).Parse(page)
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
	data.Transcripts.Segments = append(data.Transcripts.Segments[:jsonData.IDb], data.Transcripts.Segments[jsonData.IDb+1:]...)
	for i, _ := range data.Transcripts.Segments {
		data.Transcripts.Segments[i].ID = i
	}
	data.WriteJSON()
	tmpl, err := template.New("page").Funcs(tmlpFuncs).Parse(page)
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

func (c *App) Editor(uuid string, segmentID int) revel.Result {
	data := fetchJSON(uuid)
	if data.UUID == "" {
		return c.Redirect("/")
	}
	return c.Render(data, segmentID)
}

func (c *App) PlayAudio(uuid string) revel.Result {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.mp3", uuid))
	return c.RenderFile(audio, revel.Attachment)
}

func (c *App) Correct(uuid string) revel.Result {
	data := fetchJSON(uuid)
	return c.Render(data)
}

func (c *App) DeleteRecord(uuid string) revel.Result {
	base := "/data/recordings/%s.%s"
	err := os.Remove(fmt.Sprintf(base, uuid, "json"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, uuid, "mp3"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, uuid, "wav"))
	if err != nil {
		revel.AppLog.Error("Failed to DELETE", err)
	}
	err = os.Remove(fmt.Sprintf(base, uuid, "backup.json"))
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
