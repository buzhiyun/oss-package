package progress

import (
	"github.com/gosuri/uiprogress"
)

func NewProgressBar(total int64) *uiprogress.Bar {
	bar := uiprogress.AddBar(int(total)).AppendCompleted().PrependElapsed()
	bar.Set(0)
	return bar
}

func EnableProgressBar() {
	uiprogress.Start()
}
