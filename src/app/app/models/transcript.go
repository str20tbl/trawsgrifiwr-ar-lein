package models

import (
	"fmt"
	"github.com/revel/revel"
	"math"
	"os"
	"time"
)

type Transcript struct {
	ID       string    `json:"id"`
	Version  int       `json:"version"`
	Success  bool      `json:"success"`
	Segments []Segment `json:"segments"`
}

type Segment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// MergeSegments merge the segments IDa & IDb into a single segment and re index the entire list
func (t *Transcript) MergeSegments(IDa, IDb int) {
	tID := IDa
	if IDb < IDa {
		IDa = IDb
		IDb = tID
	}
	// set the end time to the end time of IDb
	t.Segments[IDa].End = t.Segments[IDb].End
	// concatenate the text from both segments
	t.Segments[IDa].Text += " " + t.Segments[IDb].Text
	// remove the segment at IDb
	t.Segments = append(t.Segments[:IDb], t.Segments[IDb+1:]...)
	// reindex the entire list
	for i := range t.Segments {
		t.Segments[i].ID = i
	}
}

// ExportSRT convert the segments into an SRT format
func (t *Transcript) ExportSRT(uid string) (filename string) {
	// a string to represent the SRT data
	srtData := ""
	for _, el := range t.Segments {
		// convert times to SRT standard "T15:04:05.999Z"
		startTime := time.Duration(el.Start * float64(time.Second))
		start := time.Unix(0, 0).UTC().Add(startTime).Format("T15:04:05.999Z")
		endTime := time.Duration(el.End * float64(time.Second))
		end := time.Unix(0, 0).UTC().Add(endTime).Format("T15:04:05.999Z")
		// build SRT data for each segment
		srtData += fmt.Sprintf("%d\n%s --> %s\n%s\n\n", el.ID, start, end, el.Text)
	}
	// write file to disk and return filename to controller for rendering
	filename = fmt.Sprintf("/data/recordings/%s.srt", uid)
	if err := os.WriteFile(filename, []byte(srtData), 0666); err != nil {
		revel.AppLog.Error("unable to write SRT", err)
	}
	return
}

const timeBuffer = 0.1
const timePrecision = 3

// NewSegment insert a new segment after the current ID unless there is less than 1 second between ID & ID + 1
func (t *Transcript) NewSegment(ID int, duration float64) {
	// assume that we are appending to the end of the array
	endTime := duration
	// if in middle then end time is start time of ID + 1 minus timeBuffer
	if ID < len(t.Segments)-1 {
		endTime = t.Segments[ID+1].Start - timeBuffer
	}
	// start time will always be end time of ID plus timeBuffer
	startTime := t.Segments[ID].End + timeBuffer
	// check we have a greater than 1 second time gap between segments
	if endTime-startTime > 1 {
		newSeg := Segment{
			ID:    0,
			Start: roundFloat(startTime, timePrecision),
			End:   roundFloat(endTime, timePrecision),
			Text:  "",
		}
		// create a new list containing newSeg at correct index
		tList := make([]Segment, 0)
		for i, el := range t.Segments {
			tList = append(tList, el)
			// insert after the curent ID
			if i == ID {
				tList = append(tList, newSeg)
			}
		}
		t.Segments = tList
	}
	// reindex the segments
	for i := range t.Segments {
		t.Segments[i].ID = i
	}
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
