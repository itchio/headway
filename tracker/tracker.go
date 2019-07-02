package tracker

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/itchio/headway/ewma"
	"github.com/itchio/headway/united"
)

// OnFinish is the callback type for tracker finition events
type OnFinish func()

// A Tracker tracks the progress of a task, and estimates
// time left, bytes per second (if relevant)
type Tracker interface {
	// Pause temporarily stops progress tracking (resets speed / time left)
	Pause()
	// Resume restarts progress tracking (resets speed / time left)
	Resume()
	// Paused returns true if this tracker is temporarily paused
	Paused() bool
	// OnFinish registers a finish callback for this tracker
	OnFinish(callback OnFinish)

	// ByteAmount returns the amount of bytes the task this tracker tracks has to go through (if relevant)
	ByteAmount() *ByteAmount

	// SetProgress sets the current value. Setting to a lower value than the current value resets speed & time left
	SetProgress(value float64)
	// Duration returns the amount of time spent tracking progress (excluding pauses)
	Duration() time.Duration
	// Progress returns the current progress value
	Progress() float64

	// Stats returns speed & time left, if they're accurate enough
	Stats() *Stats

	Finish() CompletionStats
}

// ByteAmount represents an amount in bytes
type ByteAmount struct {
	Value int64
}

func (ba ByteAmount) String() string {
	return united.FormatBytes(ba.Value)
}

type tracker struct {
	startTime           time.Time
	value               float64
	max                 float64
	measurementInterval time.Duration
	paused              bool

	onFinish []OnFinish

	mutex    sync.Mutex
	duration time.Duration

	speed              float64
	minSpeed           float64
	maxSpeed           float64
	speedAverage       ewma.Average
	secondsLeftAverage ewma.Average
	lastMeasurement    *measurement

	byteAmount *ByteAmount
}

type measurement struct {
	time  time.Time
	value float64
}

var _ Tracker = (*tracker)(nil)

// CompletionStats contains statistics on the duration and speed of a task
// tracked with a tracker
type CompletionStats struct {
	duration     time.Duration
	averageSpeed float64
	minSpeed     float64
	maxSpeed     float64
	byteAmount   *ByteAmount
}

func (cs CompletionStats) String() string {
	return fmt.Sprintf("(%v total, avg %.2f/sec, min %.2f/sec, max %.2f/sec)", cs.duration, cs.averageSpeed, cs.minSpeed, cs.maxSpeed)
}

// Duration returns how long the task was tracked for (excluding pauses)
func (cs CompletionStats) Duration() time.Duration {
	return cs.duration
}

// ByteAmount returns the byte amount associated with the task, if any
func (cs CompletionStats) ByteAmount() *ByteAmount {
	return cs.byteAmount
}

// AverageSpeed returns an average of the speed the tracker recorded
func (cs CompletionStats) AverageSpeed() float64 {
	return cs.averageSpeed
}

// AverageBPS returns the average bandwidth (if a byte amount was set)
func (cs CompletionStats) AverageBPS() *BPS {
	return toBPS(cs.byteAmount, cs.averageSpeed)
}

// MinSpeed returns the lowest speed the tracker recorded
func (cs CompletionStats) MinSpeed() float64 {
	return cs.minSpeed
}

// MaxSpeed returns the highest speed the tracker recorded
func (cs CompletionStats) MaxSpeed() float64 {
	return cs.maxSpeed
}

// BPS represents an amount of bytes per second
type BPS struct {
	Value float64
}

func (bps BPS) String() string {
	return united.FormatBPSValue(bps.Value)
}

// Stats can be queried at any moment during a tracker's timelife,
// but may be unavailable if not enough data has been fed yet.
type Stats struct {
	value float64

	// Speed is the progress speed measured in the last interval
	speed float64

	// TimeLeft represents the amount of time after which tracker believes the task will be finished,
	// if it keeps at its current average speed.
	timeLeft *time.Duration

	byteAmount *ByteAmount
}

// Value returns the current progress of the task
func (s Stats) Value() float64 {
	return s.value
}

// Speed returns the current speed of the task, without units
func (s Stats) Speed() float64 {
	return s.speed
}

// TimeLeft returns an estimate of how long it will take to complete the task.
func (s Stats) TimeLeft() *time.Duration {
	return s.timeLeft
}

// BPS returns a bandwidth, only if the task has an associated byte amount
func (s Stats) BPS() *BPS {
	return toBPS(s.byteAmount, s.speed)
}

// ByteAmount returns the byte amount for this task (if any)
func (s Stats) ByteAmount() *ByteAmount {
	return s.byteAmount
}

func (s Stats) String() string {
	speed := fmt.Sprintf("%.2f/sec", s.speed)
	left := "unknown time left"
	if s.timeLeft != nil {
		left = fmt.Sprintf("%v left", s.timeLeft)
	}
	return fmt.Sprintf("(%.2f%% done @ %s, %s)", s.value*100.0, speed, left)
}

// Opts configures a tracker
type Opts struct {
	ByteAmount          *ByteAmount
	Value               float64
	Units               united.Units
	MeasurementInterval time.Duration
}

func (opts *Opts) ensureDefaults() {
	var zero time.Duration
	if opts.MeasurementInterval == zero {
		opts.MeasurementInterval = 1 * time.Second
	}
}

// New creates a new tracker and starts it
func New(opts Opts) Tracker {
	opts.ensureDefaults()

	t := &tracker{
		startTime:           time.Now(),
		value:               opts.Value,
		measurementInterval: opts.MeasurementInterval,
		byteAmount:          opts.ByteAmount,

		speed:              0,
		minSpeed:           math.MaxFloat64,
		maxSpeed:           0,
		speedAverage:       ewma.New(0),
		secondsLeftAverage: ewma.New(0),
	}
	return t
}

func (t *tracker) Finish() CompletionStats {
	for _, cb := range t.onFinish {
		cb()
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.lastMeasurement != nil {
		t.duration += time.Since(t.lastMeasurement.time)
		t.lastMeasurement = nil
	}

	return CompletionStats{
		duration:     t.duration,
		averageSpeed: 1.0 / t.duration.Seconds(),
		minSpeed:     t.minSpeed,
		maxSpeed:     t.maxSpeed,
		byteAmount:   t.byteAmount,
	}
}

func (t *tracker) Pause() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.paused = true
	t.lockedResetMeasurement()
}

func (t *tracker) Resume() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.paused = false
	t.lockedResetMeasurement()
}

func (t *tracker) SetProgress(value float64) {
	value = clamp(value)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.lockedUpdateMeasurement(value)
	t.value = value
}

// must hold mutex
func (t *tracker) lockedUpdateMeasurement(value float64) {
	if t.paused {
		t.lockedResetMeasurement()
	}

	now := time.Now()

	lastMeasurement := t.lastMeasurement
	if lastMeasurement == nil {
		t.lastMeasurement = &measurement{
			time:  now,
			value: value,
		}
		return
	}

	sinceLast := time.Since(lastMeasurement.time)
	if sinceLast < t.measurementInterval {
		// don't update yet
		return
	}
	t.duration += sinceLast

	valueDelta := value - lastMeasurement.value
	if valueDelta < 0 {
		// went back in the past huh?
		// in this case, reset everything
		t.lockedResetMeasurement()
		return
	}

	t.speed = valueDelta / sinceLast.Seconds()
	t.speedAverage.Add(t.speed)

	if t.speed > t.maxSpeed {
		t.maxSpeed = t.speed
	}
	if t.speed < t.minSpeed {
		t.minSpeed = t.speed
	}

	{
		secondsLeft := (1.0 - t.value) / t.speedAverage.Value()
		t.secondsLeftAverage.Add(secondsLeft)
	}

	t.lastMeasurement = &measurement{
		time:  now,
		value: value,
	}
}

// must hold mutex
func (t *tracker) lockedResetMeasurement() {
	t.lastMeasurement = nil
	t.speed = 0
	t.minSpeed = math.MaxFloat64
	t.maxSpeed = 0
	t.speedAverage = ewma.New(0)
	t.secondsLeftAverage = ewma.New(0)
}

func (t *tracker) Progress() float64 {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.value
}

func (t *tracker) Duration() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.duration
}

func (t *tracker) Stats() *Stats {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.speed == 0.0 || t.lastMeasurement == nil {
		return nil
	}

	secondsLeft := (1.0 - t.value) / t.speedAverage.Value()
	timeLeftVal := time.Millisecond * time.Duration(secondsLeft*1000.0)
	timeLeft := &timeLeftVal
	if timeLeftVal < time.Duration(0) {
		timeLeft = nil
	}

	return &Stats{
		speed:      t.speedAverage.Value(),
		timeLeft:   timeLeft,
		value:      t.value,
		byteAmount: t.byteAmount,
	}
}

func (t *tracker) Paused() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.paused
}

func (t *tracker) OnFinish(callback OnFinish) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.onFinish = append(t.onFinish, callback)
}

func (t *tracker) ByteAmount() *ByteAmount {
	return t.byteAmount
}

func clamp(value float64) float64 {
	if value > 1.0 {
		return 1.0
	} else if value < 0.0 {
		return 0.0
	}
	return value
}

func toBPS(byteAmount *ByteAmount, speed float64) *BPS {
	if byteAmount == nil {
		return nil
	}

	return &BPS{
		Value: speed * float64(byteAmount.Value),
	}
}
