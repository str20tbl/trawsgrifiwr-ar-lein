package controllers

import (
	"app/app/appJobs"
	"app/app/models"
	"bytes"
	"fmt"
	"github.com/revel/revel"
	"html/template"
	"time"
)

func renderTemplate(data appJobs.ProcessFiles) (output bytes.Buffer) {
	tmpl, err := template.New("page").Funcs(tmlpFuncs).Parse(page)
	if err != nil {
		revel.AppLog.Error("Parse", err)
	}
	err = tmpl.Execute(&output, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		revel.AppLog.Error("Execute", err)
	}
	return
}

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
	"isIntern": func(input string, blackList []string) (fnd bool) {
		for _, item := range blackList {
			if input == item {
				fnd = true
				break
			}
		}
		return fnd
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
                                            <div class="col-sm-12 pb-3 ">
                                                <textarea class="form-control" id="staticEmail2" rows="7" onfocusout="saveFile()">{{.Text}}</textarea>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="card-footer">
                                        <div class="row">
                                            <div class="col-12 py-3">
                                                <button class="w-100 input-group-text btn btn-sm btn-outline-primary" onclick="addSegment({{.ID}})">
                                                    + <span class="en">New</span> Segment <span class="cy">Newydd</span>
                                                </button>
                                            </div>
                                            <div class="col-6">
                                                {{if gt .ID 0}}
                                                <button class="w-100 btn btn-sm btn-outline-primary" onclick="mergeSegments({{.ID}}, {{add .ID -1}})">
                                                    <span class="cy">Cyfuno gyda</span>
                                                    <span class="en">Merge with</span>
                                                    ID: {{add .ID -1}}
                                                </button>
                                                {{else}}
                                                &nbsp
                                                {{end}}
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
