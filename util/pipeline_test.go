package util

import (
	"testing"
)

func Test(t *testing.T) {
	t.Log(GetMachineID("tasks/1260036427622647809/stages/grayify-image/results/1,1260036427819786241.png"))
}
