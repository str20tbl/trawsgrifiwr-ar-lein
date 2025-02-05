package controllers

import (
	"app/app/appJobs"
	"app/app/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"html/template"
	"math"
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
	"prevTime": func(input models.Transcript, index int) float64 {
		if index == 0 {
			return 0.0
		}
		return input.Segments[index-1].Start
	},
	"nextTime": func(input models.Transcript, index int) float64 {
		if index+1 >= len(input.Segments)-1 {
			return input.Segments[len(input.Segments)-1].Start
		}
		return input.Segments[index+1].Start
	},
	"len": func(input models.Transcript) int {
		return len(input.Segments) - 1
	},
	"notIn": func(uuid string, blackList []string) (fnd bool) {
		for _, item := range blackList {
			if uuid == item {
				fnd = true
				break
			}
		}
		return !fnd
	},
}

const slider = `<input type="range" class="form-range" id="customRange"
                                       min="0" max="{{len .data.Transcripts}}" title="dewis segment">`

const page = `<div class="row">
                            <div class="col-2">
                                {{range .data.Transcripts.Segments}}
                                <div id="{{.ID}}_text_left" class="hide col-6 transcript h-100 w-100">
                                    <div class="card m-3 pt-3 h-100">
                                        <div class="card-header" style="height: 100px">
                                            <div class="row">
                                                <div class="col">
                                                    ID: {{.ID}}
                                                </div>
                                                <div class="col">
                                                    <div>
                                                        <span class="cy">Diwedd</span><span class="en">End</span>: <span>{{s2m .End}}</span>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="card-body">
                                            <span class="align-middle">{{.Text}}</span>
                                        </div>
                                    </div>
                                </div>
                                {{end}}
                            </div>
                            <div class="col-8">
                                {{range .data.Transcripts.Segments}}
                                <div class="card m-3 pt-3 row hide transcript  h-100 w-100" id="{{.ID}}_tran">
                                    <div class="card-header" style="height: 100px">
                                        <div class="row">
                                            <div class="col">
                                                <button class="btn btn-sm btn-outline-primary" title="chwarae segment o'r cychwyn" id="play-{{.ID}}" onclick="restart_segment({{.ID}})">
                                                    <svg width="18px" height="18px" xmlns="http://www.w3.org/2000/svg" fill="currentColor" class="bi bi-play-fill" viewBox="0 0 16 16">
                                                        <path d="m11.596 8.697-6.363 3.692c-.54.313-1.233-.066-1.233-.697V4.308c0-.63.692-1.01 1.233-.696l6.363 3.692a.802.802 0 0 1 0 1.393z"></path>
                                                    </svg>
                                                </button>
                                                <button class="btn btn-sm btn-outline-primary" id="repeat-{{.ID}}" title="chwarae segment mewn lŵp " onclick="loop_segment({{.ID}})">
                                                    <svg width="18px" height="18px" xmlns="http://www.w3.org/2000/svg" fill="currentColor" class="bi bi-recycle" viewBox="0 0 16 16">
                                                        <path d="M9.302 1.256a1.5 1.5 0 0 0-2.604 0l-1.704 2.98a.5.5 0 0 0 .869.497l1.703-2.981a.5.5 0 0 1 .868 0l2.54 4.444-1.256-.337a.5.5 0 1 0-.26.966l2.415.647a.5.5 0 0 0 .613-.353l.647-2.415a.5.5 0 1 0-.966-.259l-.333 1.242-2.532-4.431zM2.973 7.773l-1.255.337a.5.5 0 1 1-.26-.966l2.416-.647a.5.5 0 0 1 .612.353l.647 2.415a.5.5 0 0 1-.966.259l-.333-1.242-2.545 4.454a.5.5 0 0 0 .434.748H5a.5.5 0 0 1 0 1H1.723A1.5 1.5 0 0 1 .421 12.24l2.552-4.467zm10.89 1.463a.5.5 0 1 0-.868.496l1.716 3.004a.5.5 0 0 1-.434.748h-5.57l.647-.646a.5.5 0 1 0-.708-.707l-1.5 1.5a.498.498 0 0 0 0 .707l1.5 1.5a.5.5 0 1 0 .708-.707l-.647-.647h5.57a1.5 1.5 0 0 0 1.302-2.244l-1.716-3.004z"></path>
                                                    </svg>
                                                </button>
                                            </div>
                                            <div class="col">
                                                ID: {{.ID}}
                                            </div>
                                            <div class="col">
                                                <div>
                                                    <label for="startRange-{{.ID}}" class="form-label"><span class="cy">Dechrau</span><span class="en">Start</span>: <span id="startTime-{{.ID}}">{{s2m .Start}}</span></label>
                                                    <input type="range" class="form-range" id="startRange-{{.ID}}" step="0.001" value="{{.Start}}" min="{{prevTime $.data.Transcripts .ID}}" max="{{.End}}">
                                                </div>
                                            </div>
                                            <div class="col">
                                                <div>
                                                    <label for="endRange-{{.ID}}" class="form-label"><span class="cy">Diwedd</span><span class="en">End</span>: <span id="endTime-{{.ID}}">{{s2m .End}}</span></label>
                                                    <input type="range" class="form-range" id="endRange-{{.ID}}" step="0.001" value="{{.End}}" min="{{.Start}}" max="{{nextTime $.data.Transcripts .ID}}">
                                                </div>
                                            </div>
                                            <script>
                                                $(document).on('input', '#startRange-{{.ID}}', function() {
                                                    let value = Number($(this).val());
                                                    syncData.forEach(function (element, index, array) {
                                                        if (index === {{.ID}}) {
                                                            element.start = value;
                                                        }
                                                    });
                                                    $("#startTime-{{.ID}}").text(value + "s");
                                                    saveFile();
                                                });
                                                $(document).on('input', '#endRange-{{.ID}}', function() {
                                                    let value = Number($(this).val());
                                                    syncData.forEach(function (element, index, array) {
                                                        if (index === {{.ID}}) {
                                                            element.end = value;
                                                        }
                                                    });
                                                    $("#endTime-{{.ID}}").text(value + "s");
                                                    saveFile();
                                                });
                                            </script>
                                        </div>
                                    </div>
                                    <div class="card-body">
                                        <div class="row">
                                            <div class="col-sm-12 pb-3">
                                                <textarea class="form-control" id="staticEmail2" rows="5" onfocusout="saveFile()">{{.Text}}</textarea>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="card-footer">
                                        <div class="row">
                                            <div class="col-6">
                                                <button class="w-100 input-group-text btn btn-sm btn-outline-primary" onclick="addSegment({{.ID}})">
                                                    + <span class="en">New</span> Segment <span class="cy">Newydd</span>
                                                </button>
                                            </div>
                                            <div class="col-6">
                                                {{ $length := len $.data.Transcripts }} {{if lt .ID $length}}
                                                <button class="w-100 btn btn-sm btn-outline-primary" onclick="mergeSegments({{.ID}}, {{add .ID 1}})">
                                                    <span class="cy">Cyfuno gyda</span>
                                                    <span class="en">Merge with</span>
                                                    ID: {{add .ID 1}}
                                                </button>
                                                {{else}}
                                                &nbsp
                                                {{end}}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                {{end}}
                            </div>
                            <div class="col-2">
                                {{range .data.Transcripts.Segments}}
                                <div id="{{.ID}}_text_right" class="hide col-6 transcript h-100 w-100">
                                    <div class="card m-3 pt-3 h-100">
                                        <div class="card-header" style="height: 100px">
                                            <div class="row">
                                                <div class="col">
                                                    <div>
                                                        <span class="cy">Dechrau</span><span class="en">Start</span>: <span>{{s2m .Start}}</span>
                                                    </div>
                                                </div>
                                                <div class="col">
                                                    ID: {{.ID}}
                                                </div>
                                            </div>
                                        </div>
                                        <div class="card-body">
                                            <span class="align-middle">{{.Text}}</span>
                                        </div>
                                    </div>
                                </div>
                                {{end}}
                            </div>
                        </div>`

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
	data := fetchJSON(jsonData.UUID)
	data.Transcripts.NewSegment(jsonData.ID, data.Duration)
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

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
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
	data := fetchJSON(jsonData.UUID)
	data.Transcripts.MergeSegments(jsonData.IDa, jsonData.IDb)
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

func (c *App) GetData(id string) revel.Result {
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
	srtFilename := data.Transcripts.ExportSRT(uuid)
	return c.RenderFileName(srtFilename, revel.Attachment)
}

func (c *App) UpdateJSON() revel.Result {
	var jsonData struct {
		UUID string            `json:"uuid"`
		Data models.Transcript `json:"data"`
	}
	err := c.Params.BindJSON(&jsonData)
	if err != nil {
		revel.AppLog.Errorf("Unable to bind JSON :: %v", err)
	}
	uid := fmt.Sprintf("%s", jsonData.UUID)
	plan, _ := os.ReadFile(fmt.Sprintf("/data/recordings/%s.json", uid))
	var originalJSON appJobs.ProcessFiles
	err = json.Unmarshal(plan, &originalJSON)
	if err != nil {
		revel.AppLog.Error("Unable to open JSON")
	}
	originalJSON.Transcripts = jsonData.Data
	originalJSON.Updated = time.Now().Format("2006-01-02 15:04")
	if originalJSON.Started == "0001-01-01T00:00:00Z" || originalJSON.Started == "" {
		originalJSON.Started = time.Now().Format("2006-01-02 15:04")
	}
	originalJSON.WriteJSON()
	return c.RenderJSON(`{"success": "True"}`)
}

var UUIDBlackList = []string{
	"d32a21ff-b1a3-4e92-b5f4-6f81f9e36bc8",
	"5b878f9b-6ede-440d-8393-0f38b19645df",
	"e44e1367-561b-482b-bead-876a9b841631",
	"09a9b06d-e014-4ee2-bdf0-5646440b2afd",
	"b64c90d4-67ad-4be6-90d7-b0f15b6d09ed",
	"cc17423a-65b7-46c8-947e-b3eb44fb856a",
	"73566209-4245-417b-970e-da24eb832f05",
	"6a9e26f9-ed5f-47a3-976b-66b39d1e7251",
	"92f82aa6-4c2c-4819-a612-9bbf1d5fa618",
	"8c68b86c-9031-4817-8d49-071b6f18cff5",
	"1280b37c-8cde-4d32-8492-9004eb7620f2",
	"e740cf15-b39a-4480-a4a2-b06de0c5388d",
	"3134ca38-67ad-4440-9ab1-80931bed1387",
	"3264c16b-c238-44a9-a282-8aa88d45516f",
	"b89330c3-c647-4b39-83de-e3531c722aa9",
	"48450467-0286-4f05-bbf3-e6430b0e454a",
	"ffcfdee7-f5d2-41cd-b736-a6dc10d24552",
	"b0ffe67b-d555-47d2-9a8f-6ed5f74c4861",
	"cffce654-4a9d-45db-a665-c3c0e19cf955",
	"427731ed-7405-470b-b7ea-5f3570aefecd",
	"9125c49a-581a-4a54-95ea-0bb51d49f3ac",
	"21965db8-0a02-41e5-bb56-d42b8cc9fec0",
	"025cb6e0-8ec2-416b-a161-40a363ec9624",
	"b7c62ff2-8af9-44ab-ba0c-752ce60511d0",
	"8a40370f-d091-4bd1-b42d-e4dcadd9d028",
	"78680142-3dcf-4f3b-ac39-9c6e1341d0c7",
	"f8518996-89a3-4892-a5c8-b0ad3408b47a",
	"acce4f34-6b3b-449c-8bdc-d1e2146d08d2",
	"f3232989-d321-44a8-8530-fcaec79bec9d",
	"e937a3c8-6665-4fa7-b59f-04ee6c3d06b6",
	"f0f5bdee-4f63-4f26-9332-f7f2cd0bd88f",
	"0e46bb4a-f114-4367-8b61-dc93aa704440",
	"de68c6e4-34d1-47eb-9132-6f43c1fe4df7",
	"8c68b86c-9031-4817-8d49-071b6f18cffo",
}

func (c *App) Editor(uuid string, segmentID int) revel.Result {
	data := fetchJSON(uuid)
	if data.UUID == "" {
		return c.Redirect("/")
	}
	return c.Render(data, segmentID, UUIDBlackList)
}

func (c *App) PlayAudio(uuid string) revel.Result {
	audio, _ := os.Open(fmt.Sprintf("/data/recordings/%s.wav", uuid))
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
	job.WriteJSON()
	jobs.Now(job)

	return c.Redirect((*App).Correct, fileUUID.String())
}
