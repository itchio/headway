package tracker_test

import (
	"testing"
	"time"

	"github.com/itchio/headway/tracker"
	"github.com/stretchr/testify/assert"
)

func Test_TrackerConstant(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	var lastStats *tracker.Stats

	for f := 0.0; f <= 1.0; f += 0.1 {
		time.Sleep(10 * time.Millisecond)
		tr.SetProgress(f)

		stats := tr.Stats()
		if lastStats != nil && stats != nil {
			assert.Less(stats.TimeLeft().Seconds(), lastStats.TimeLeft().Seconds())
			assert.InEpsilon(stats.Speed(), lastStats.Speed(), 0.2)
		}

		t.Logf("%v", tr.Stats())
		lastStats = stats
	}

	stats := tr.Finish()
	t.Logf("%v", stats)

	assert.InEpsilon(0.1, stats.Duration().Seconds(), 0.15)
	assert.InEpsilon(stats.MinSpeed(), stats.AverageSpeed(), 0.15)
	assert.InEpsilon(stats.MaxSpeed(), stats.AverageSpeed(), 0.15)
}

func Test_TrackerRampUp(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	var lastStats *tracker.Stats

	speed := 0.01
	progress := 0.0
	for {
		time.Sleep(10 * time.Millisecond)
		speed *= 1.05
		progress += speed

		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		stats := tr.Stats()
		if lastStats != nil && stats != nil {
			assert.Less(stats.TimeLeft().Seconds(), lastStats.TimeLeft().Seconds())
			assert.GreaterOrEqual(stats.Speed(), lastStats.Speed())
		}

		t.Logf("%v", tr.Stats())
		lastStats = stats
	}

	stats := tr.Finish()
	t.Logf("%v", stats)
}

func Test_TrackerRampDown(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	var lastStats *tracker.Stats

	speed := 0.1
	progress := 0.0
	for {
		time.Sleep(10 * time.Millisecond)
		speed *= 0.93
		progress += speed

		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		stats := tr.Stats()
		if lastStats != nil && stats != nil {
			assert.Less(stats.TimeLeft().Seconds(), lastStats.TimeLeft().Seconds())
			assert.LessOrEqual(stats.Speed(), lastStats.Speed())
		}

		t.Logf("%v", tr.Stats())
		lastStats = stats
	}

	stats := tr.Finish()
	t.Logf("%v", stats)
}

func Test_TrackerRampUpAndDown(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	var lastStats *tracker.Stats

	speed := 0.01
	progress := 0.0

	delayRounds := 5

	for {
		time.Sleep(10 * time.Millisecond)
		rampingUp := progress < 0.5
		if rampingUp {
			speed *= 1.1
		} else {
			speed *= 0.93
		}
		progress += speed

		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		stats := tr.Stats()
		if lastStats != nil && stats != nil {
			assert.Less(stats.TimeLeft().Seconds(), lastStats.TimeLeft().Seconds())
			if rampingUp {
				assert.GreaterOrEqual(stats.Speed(), lastStats.Speed())
			} else {
				if delayRounds > 0 {
					delayRounds--
				} else {
					assert.LessOrEqual(stats.Speed(), lastStats.Speed())
				}
			}
		}

		t.Logf("%v", tr.Stats())
		lastStats = stats
	}

	stats := tr.Finish()
	t.Logf("%v", stats)
}

func Test_TrackerBrutalHalving(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	speed := 0.01
	progress := 0.0

	for {
		time.Sleep(10 * time.Millisecond)
		halved := progress >= 0.5
		if halved {
			speed = 0.005
		}
		progress += speed

		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		stats := tr.Stats()
		t.Logf("%v", stats)

		if stats != nil {
			assert.GreaterOrEqual(stats.Speed(), 0.4)
			assert.LessOrEqual(stats.Speed(), 1.0)

			if halved {
				if progress > 0.75 {
					assert.InEpsilon(0.5, stats.Speed(), 0.15)
				}
			} else {
				assert.InEpsilon(1.0, stats.Speed(), 0.15)
			}
		}
	}

	stats := tr.Finish()
	t.Logf("%v", stats)
}

func Test_TrackerJigsaw(t *testing.T) {
	assert := assert.New(t)

	tr := tracker.New(tracker.Opts{
		MeasurementInterval: 1 * time.Millisecond,
	})

	fast := true
	iters := 0

	speed := 0.01
	progress := 0.0

	for {
		time.Sleep(10 * time.Millisecond)

		iters++
		if iters > 10 {
			fast = !fast
		}

		if fast {
			speed = 0.01
		} else {
			speed = 0.001
		}
		progress += speed

		if progress > 1.0 {
			tr.SetProgress(1.0)
			break
		} else {
			tr.SetProgress(progress)
		}

		stats := tr.Stats()
		t.Logf("%v", stats)
	}

	stats := tr.Finish()
	t.Logf("%v", stats)
	assert.InEpsilon(0.5, stats.AverageSpeed(), 0.3)
	assert.InEpsilon(0.1, stats.MinSpeed(), 0.2)
	assert.InEpsilon(1, stats.MaxSpeed(), 0.2)
}
