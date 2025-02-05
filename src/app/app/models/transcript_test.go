package models

import (
	"testing"
)

func TestTranscript_ExportSRT(t1 *testing.T) {
	type fields struct {
		ID       string
		Version  int
		Success  bool
		Segments []Segment
	}
	type args struct {
		uid string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFilename string
	}{
		{
			name: "Test Write File",
			fields: fields{
				ID:       "",
				Version:  0,
				Success:  true,
				Segments: testSegments,
			},
			args: args{
				uid: "uuid-test-filename",
			},
			wantFilename: "/data/recordings/uuid-test-filename.srt",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transcript{
				ID:       tt.fields.ID,
				Version:  tt.fields.Version,
				Success:  tt.fields.Success,
				Segments: tt.fields.Segments,
			}
			if gotFilename := t.ExportSRT(tt.args.uid); gotFilename != tt.wantFilename {
				t1.Errorf("ExportSRT() = %v, want %v", gotFilename, tt.wantFilename)
			}
		})
	}
}

func TestTranscript_MergeSegments(t1 *testing.T) {
	type fields struct {
		ID       string
		Version  int
		Success  bool
		Segments []Segment
	}
	type args struct {
		IDa int
		IDb int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		testLength int
	}{
		{
			name: "Test large list start",
			fields: fields{
				ID:       "",
				Version:  0,
				Success:  true,
				Segments: testSegments,
			},
			args: args{
				IDa: 0,
				IDb: 1,
			},
			testLength: len(testSegments) - 1,
		},
		{
			name: "Test large list end",
			fields: fields{
				ID:       "",
				Version:  0,
				Success:  true,
				Segments: testSegments,
			},
			args: args{
				IDa: 175,
				IDb: 176,
			},
			testLength: len(testSegments) - 1,
		},
		{
			name: "Test large list middle",
			fields: fields{
				ID:       "",
				Version:  0,
				Success:  true,
				Segments: testSegments,
			},
			args: args{
				IDa: 50,
				IDb: 51,
			},
			testLength: len(testSegments) - 1,
		},
		{
			name: "Test small list content",
			fields: fields{
				ID:       "",
				Version:  0,
				Success:  true,
				Segments: simpleTestSegments,
			},
			args: args{
				IDa: 0,
				IDb: 1,
			},
			testLength: len(simpleTestSegments) - 1,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transcript{
				ID:       tt.fields.ID,
				Version:  tt.fields.Version,
				Success:  tt.fields.Success,
				Segments: tt.fields.Segments,
			}
			if t.MergeSegments(tt.args.IDa, tt.args.IDb); len(t.Segments) != tt.testLength {
				t1.Errorf("MergeSegments() = %v, want %v", len(t.Segments), tt.testLength)
			}
		})
	}
}

func TestTranscript_MergeSegments_Content(t1 *testing.T) {
	type fields struct {
		ID       string
		Version  int
		Success  bool
		Segments []Segment
	}
	type args struct {
		IDa int
		IDb int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantText string
	}{
		{
			name: "Test small list content",
			fields: fields{
				ID:      "",
				Version: 0,
				Success: true,
				Segments: []Segment{
					{
						ID:    0,
						Start: 0.009,
						End:   7.876,
						Text:  "a",
					},
					{
						ID:    1,
						Start: 8.116,
						End:   16.364,
						Text:  "b",
					},
				},
			},
			args: args{
				IDa: 0,
				IDb: 1,
			},
			wantText: "a b",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transcript{
				ID:       tt.fields.ID,
				Version:  tt.fields.Version,
				Success:  tt.fields.Success,
				Segments: tt.fields.Segments,
			}
			if t.MergeSegments(tt.args.IDa, tt.args.IDb); t.Segments[tt.args.IDa].Text != tt.wantText {
				t1.Errorf("MergeSegments() = %v, want %v", t.Segments[tt.args.IDa].Text, tt.wantText)
			}
		})
	}
}

func TestTranscript_NewSegment(t1 *testing.T) {
	type fields struct {
		ID       string
		Version  int
		Success  bool
		Segments []Segment
	}
	type args struct {
		ID       int
		duration float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLen int
	}{
		{
			name: "TestFile",
			fields: fields{
				ID:      "",
				Version: 0,
				Success: true,
				Segments: []Segment{
					{
						ID:    0,
						Start: 0.009,
						End:   7.876,
						Text:  "a",
					},
					{
						ID:    1,
						Start: 8.116,
						End:   16.364,
						Text:  "b",
					},
				},
			},
			args: args{
				ID:       1,
				duration: 20,
			},
			wantLen: 3,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transcript{
				ID:       tt.fields.ID,
				Version:  tt.fields.Version,
				Success:  tt.fields.Success,
				Segments: tt.fields.Segments,
			}
			if t.NewSegment(tt.args.ID, tt.args.duration); len(t.Segments) != tt.wantLen {
				t1.Errorf("MergeSegments() = %v, want %v", len(t.Segments), tt.wantLen)
			}
		})
	}
}

func Test_roundFloat(t *testing.T) {
	type args struct {
		val       float64
		precision uint
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Test floats round",
			args: args{
				val:       12.3465,
				precision: 3,
			},
			want: 12.347,
		},
		{
			name: "Test floats no round",
			args: args{
				val:       12.00000000,
				precision: 3,
			},
			want: 12.000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundFloat(tt.args.val, tt.args.precision); got != tt.want {
				t.Errorf("roundFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}
