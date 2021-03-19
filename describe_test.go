package p4

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type describeInput struct {
	args []string
	res  []map[interface{}]interface{}
}

type describeTest struct {
	input describeInput
	want  Describe
}

var describeTests = []describeTest{
	{
		input: describeInput{},
		want:  Describe{},
	},
	{
		// Single user protection
		input: describeInput{
			args: []string{"-s", "123"},
			res: []map[interface{}]interface{}{{
				"change":     "123",
				"changeType": "public",
				"client":     "user_ws",
				"code":       "stat",
				"desc":       "a description",
				"oldChange":  "122",
				"path":       "//path/to/*",
				"status":     "submitted",
				"time":       "1612369118",
				"user":       "a.person",
				"job0":       "JOB_123",
				"jobstat0":   "Reported",
				"action0":    "edit",
				"rev0":       "16",
				"depotFile0": "//path/to/file.sv",
				"type0":      "text",
				"digest0":    "71488D5623A97858A5683140F6EEF5E2",
				"fileSize0":  "21567",
			}},
		},
		want: Describe{
			Code:       "stat",
			Change:     123,
			OldChange:  122,
			ChangeType: "public",
			Client:     "user_ws",
			Desc:       "a description",
			Path:       "//path/to/*",
			Time:       time.Unix(1612369118, 0),
			Status:     "submitted",
			User:       "a.person",
			Jobs: []JobDescription{
				{
					Job:    "JOB_123",
					Status: "Reported",
				},
			},
			Revisions: []Revision{
				{
					Action:    "edit",
					Rev:       16,
					DepotFile: "//path/to/file.sv",
					Type:      "text",
					Digest:    "71488D5623A97858A5683140F6EEF5E2",
					FileSize:  21567,
				},
			},
		},
	},
	{
		// Single user protection
		input: describeInput{
			args: []string{"-s", "123"},
			res: []map[interface{}]interface{}{{
				"change":     "123",
				"changeType": "public",
				"client":     "user_ws",
				"code":       "stat",
				"desc":       "a description",
				"oldChange":  "122",
				"path":       "//path/to/*",
				"status":     "submitted",
				"time":       "1612369118",
				"user":       "a.person",
				"job0":       "JOB_123",
				"jobstat0":   "First",
				"job1":       "JOB_124",
				"jobstat1":   "Second",
				"job2":       "JOB_125",
				"jobstat2":   "Third",
				"action0":    "edit",
				"rev0":       "16",
				"depotFile0": "//path/to/file.sv",
				"type0":      "text",
				"digest0":    "71488D5623A97858A5683140F6EEF5E2",
				"fileSize0":  "21567",
				"action1":    "add",
				"rev1":       "17",
				"depotFile1": "//path/to/file2.sv",
				"type1":      "text",
				"digest1":    "71488D5623A97858A5683140F6EEF5E3",
				"fileSize1":  "21567",
				"action2":    "delete",
				"rev2":       "18",
				"depotFile2": "//path/to/file3.sv",
				"type2":      "text",
				"digest2":    "71488D5623A97858A5683140F6EEF5E4",
				"fileSize2":  "21567",
			}},
		},
		want: Describe{
			Code:       "stat",
			Change:     123,
			OldChange:  122,
			ChangeType: "public",
			Client:     "user_ws",
			Desc:       "a description",
			Path:       "//path/to/*",
			Time:       time.Unix(1612369118, 0),
			Status:     "submitted",
			User:       "a.person",
			Jobs: []JobDescription{
				{
					Job:    "JOB_123",
					Status: "First",
				},
				{
					Job:    "JOB_124",
					Status: "Second",
				},
				{
					Job:    "JOB_125",
					Status: "Third",
				},
			},
			Revisions: []Revision{
				{
					Action:    "edit",
					Rev:       16,
					DepotFile: "//path/to/file.sv",
					Type:      "text",
					Digest:    "71488D5623A97858A5683140F6EEF5E2",
					FileSize:  21567,
				},
				{
					Action:    "add",
					Rev:       17,
					DepotFile: "//path/to/file2.sv",
					Type:      "text",
					Digest:    "71488D5623A97858A5683140F6EEF5E3",
					FileSize:  21567,
				},
				{
					Action:    "delete",
					Rev:       18,
					DepotFile: "//path/to/file3.sv",
					Type:      "text",
					Digest:    "71488D5623A97858A5683140F6EEF5E4",
					FileSize:  21567,
				},
			},
		},
	},
}

func TestDescribe(t *testing.T) {
	for _, tst := range describeTests {
		fp4 := FakeP4Runner{}
		fp4.On("Run", append([]string{"describe"}, tst.input.args...)).Return(tst.input.res, nil)
		fs, err := RunDescribe(&fp4, tst.input.args)
		assert.Nil(t, err)
		assert.Equal(t, tst.want, fs)
	}
}
