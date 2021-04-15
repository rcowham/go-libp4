package p4

import "github.com/stretchr/testify/mock"

// FakeP4Runner is a mockable P4 Runner
type FakeP4Runner struct {
	mock.Mock
}

// Run Mocks p4.Run, so we can run fake perforce commands
func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}
