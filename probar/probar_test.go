package probar_test

import (
	"time"

	"github.com/itchio/headway/probar"
	"github.com/itchio/headway/tracker"
)

func ExampleBar() {
	tr := tracker.New(tracker.Opts{})
	probar.New(tr, probar.Opts{
		RefreshRate: 20 * time.Millisecond,
	})

	for f := 0.0; f < 1.0; f += 0.05 {
		time.Sleep(30 * time.Millisecond)
		tr.SetProgress(f)
	}

	tr.Finish()
}

