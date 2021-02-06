package cmds

import (
	p4 "github.com/brettbates/p4go"
	"github.com/stretchr/testify/mock"
)

type FakeP4Runner struct {
	mock.Mock
}

// Mocks p4.Run, so we can run fake perforce commands
func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}

func (p4.P4) TestFixes()
