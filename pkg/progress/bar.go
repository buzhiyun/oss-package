package progress

import (
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func NewProgressBar(total int, name ...string) *uiprogress.Bar {
	bar := uiprogress.AddBar(total).AppendCompleted().PrependElapsed()
	bar.Set(0)
	if len(name) > 0 {
		bar.PrependFunc(func(b *uiprogress.Bar) string {
			return strutil.Resize(name[0]+": ", 18)
		})
		bar.Width = 50
	}
	return bar
}

func EnableProgressBar() {
	uiprogress.Start()
}
