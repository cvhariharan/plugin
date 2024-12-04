package shared

import (
	"fmt"

	"github.com/cvhariharan/plugin"
)

type TestObj struct {
	Data string
}

func init() {
	plugin.RegisterType(TestObj{})
}

func (t *TestObj) TestCall() string {
	return fmt.Sprintf("This is a test call: %s", t.Data)
}
