package p4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fixesInput struct {
	args []string
	res  []map[interface{}]interface{}
}

type fixesTest struct {
	input fixesInput
	want  []Fix
}

var fixesTests = []fixesTest{
	{
		input: fixesInput{},
		want:  []Fix{},
	},
	{
		// Single user protection
		input: fixesInput{
			args: []string{},
			res: []map[interface{}]interface{}{{
				"Job":    "job000001",
				"Change": "1",
				"Date":   "1612571080",
				"User":   "perforce",
				"Client": "p4_ws",
				"Status": "closed",
				"code":   "stat",
			}},
		},
		want: []Fix{
			{
				Job:    "job000001",
				Change: "1",
				Date:   "1612571080",
				User:   "perforce",
				Client: "p4_ws",
				Status: "closed",
				Code:   "stat",
			},
		},
	},
	{
		// Single user protection
		input: fixesInput{
			args: []string{},
			res: []map[interface{}]interface{}{
				{
					"Job":    "job000001",
					"Change": "1",
					"Date":   "1612571080",
					"User":   "perforce",
					"Client": "p4_ws",
					"Status": "closed",
					"code":   "stat",
				},
				{
					"Job":    "job000002",
					"Change": "1",
					"Date":   "1612571081",
					"User":   "perforce",
					"Client": "p4_ws",
					"Status": "closed",
					"code":   "stat",
				},
			},
		},
		want: []Fix{
			{
				Job:    "job000001",
				Change: "1",
				Date:   "1612571080",
				User:   "perforce",
				Client: "p4_ws",
				Status: "closed",
				Code:   "stat",
			},
			{
				Job:    "job000002",
				Change: "1",
				Date:   "1612571081",
				User:   "perforce",
				Client: "p4_ws",
				Status: "closed",
				Code:   "stat",
			},
		},
	},
}

func TestFixes(t *testing.T) {
	for _, tst := range fixesTests {
		fp4 := FakeP4Runner{}
		fp4.On("Run", []string{"fixes"}).Return(tst.input.res, nil)
		fs, err := RunFixes(&fp4, tst.input.args)
		assert.Nil(t, err)
		assert.Equal(t, tst.want, fs)
	}
}
