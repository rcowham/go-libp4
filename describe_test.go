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
}

func TestDescribe(t *testing.T) {
	for _, tst := range describeTests {
		fp4 := FakeP4Runner{}
		fp4.On("Run", []string{"describe"}).Return(tst.input.res, nil)
		fs, err := RunDescribe(&fp4, tst.input.args)
		assert.Nil(t, err)
		assert.Equal(t, tst.want, fs)
	}
}
