package p4

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// FakeP4Runner is a mockable P4 Runner
type FakeP4Runner struct {
	mock.Mock
}

// Run Mocks p4.Run, so we can run fake perforce commands
func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}

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
				Change: 1,
				Date:   time.Unix(1612571080, 0),
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
