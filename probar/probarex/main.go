package main

import (
	"fmt"
	"time"

	"github.com/itchio/headway/probar"
	"github.com/itchio/headway/tracker"
	"github.com/itchio/headway/united"
)

func main() {
	tr := tracker.New(tracker.Opts{
		ByteAmount: &tracker.ByteAmount{Value: 542 * 1024 * 1024},
	})
	pb := probar.New(tr, probar.Opts{
		ShowTimeLeft: true,
		ShowSpeed:    true,
	})
	pb.SetPostfix("Fake download")

	rounds := 0
	maxrounds := 40
	dir := 1
	speed := 0.002
	progress := 0.0
	factor := 1.07

	printed := false
	printed2 := false

	for {
		rounds++
		if rounds > maxrounds {
			dir = -dir
			rounds = 0
		}
		if dir > 0 {
			speed *= factor
		} else {
			speed /= factor
		}
		progress += speed
		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		if !printed && progress >= 0.1 {
			pb.Println("Already 10% done!")
			printed = true
		}

		if !printed2 && progress >= 0.9 {
			pb.Println("Almost there!")
			printed2 = true
		}

		time.Sleep(100 * time.Millisecond)
	}

	stats := tr.Finish()
	fmt.Printf("Fake-downloaded %v in %s, @ %v on average\n", stats.ByteAmount(), united.FormatDuration(stats.Duration()), stats.AverageBPS())
}
