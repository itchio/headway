// Package probar implements a simple progress bar
package probar

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/itchio/headway/state"
	"github.com/itchio/headway/tracker"
	"github.com/itchio/headway/united"
)

// PrintFunc is the type of a function that prints a line
type PrintFunc func(f string, a ...interface{})

// Opts configures a progress bar
type Opts struct {
	RefreshRate   time.Duration
	TimeBoxWidth  int
	SpeedBoxWidth int
	BarWidth      int
	Width         int
	ShowSpeed     bool
	ShowTimeLeft  bool
	Printf        PrintFunc
}

func (opts *Opts) ensureDefaults() {
	var zero time.Duration
	if opts.RefreshRate == zero {
		opts.RefreshRate = 200 * time.Millisecond
	}
	if opts.BarWidth == 0 {
		opts.BarWidth = 20
	}
	if opts.SpeedBoxWidth == 0 {
		opts.SpeedBoxWidth = 13
	}
	if opts.TimeBoxWidth == 0 {
		opts.TimeBoxWidth = 13
	}
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.Printf == nil {
		opts.Printf = func(f string, a ...interface{}) {
			fmt.Printf(f, a...)
		}
	}
}

// Bar represents a progress bar
type Bar interface {
	// SetPrefix sets a prefix to the progress bar
	SetPrefix(prefix string)

	// SetPostfix sets a postfix to the progress bar
	SetPostfix(postfix string)

	// SetScale sets the scale of the bar. It can be scaled
	// from 0.0 to 1.0 to indicate the progress of a first task.
	SetScale(scale float64)

	// Println prints a line in a way that doesn't interfere with the
	// progress bar
	Println(s string)

	// Printfln prints a line in a way that doesn't interfere with the
	// progress bar
	Printfln(s string, a ...interface{})
}

// New creates a new progress bar tracking the given tracker
func New(tracker tracker.Tracker, opts Opts) Bar {
	opts.ensureDefaults()
	units := united.UnitsNone
	if tracker.ByteAmount() != nil {
		units = united.UnitsBytes
	}

	b := &bar{
		tracker: tracker,
		opts:    opts,
		theme:   state.GetTheme(),
		units:   units,
		scale:   1.0,

		finished:   false,
		finishChan: make(chan struct{}),
	}
	tracker.OnFinish(b.finish)
	go b.writer()
	return b
}

type bar struct {
	tracker tracker.Tracker
	opts    Opts
	theme   *state.ProgressTheme
	units   united.Units
	scale   float64

	finishChan chan struct{}
	finished   bool

	lines []string

	prefix  string
	postfix string

	mutex sync.Mutex
}

func (b *bar) SetScale(scale float64) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.scale = scale
}

// SetPrefix sets the text shown before the bar
func (b *bar) SetPrefix(prefix string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.prefix = prefix
}

// SetPostfix sets the text show after the bar
func (b *bar) SetPostfix(postfix string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.postfix = postfix
}

// Finish prints the final status of the bar and stops updating.
func (b *bar) finish() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.finished {
		return
	}

	close(b.finishChan)
	b.finished = true
	b.clear()
}

func (b *bar) Println(s string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.lines = append(b.lines, s)
}

func (b *bar) Printfln(s string, a ...interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.lines = append(b.lines, fmt.Sprintf(s, a...))
}

func (b *bar) clear() {
	b.opts.Printf("\r%s\r", strings.Repeat(" ", b.opts.Width))
}

func (b *bar) write() {
	stats := b.tracker.Stats()
	current := b.tracker.Progress()
	width := b.opts.Width

	// print lines
	if len(b.lines) > 0 {
		b.clear()
		for _, line := range b.lines {
			b.opts.Printf("%s\n", line)
		}
		b.lines = nil
	}

	var percentBox, countersBox, timeLeftBox, speedBox, barBox, end, out string
	th := b.theme

	// percents
	{
		var percent float64
		percent = current * float64(100)
		percentBox = fmt.Sprintf(" %6.02f%% ", percent)
	}

	{
		// time left
		if b.opts.ShowTimeLeft {
			if stats != nil && stats.TimeLeft() != nil {
				timeLeftBox = united.FormatDuration(*stats.TimeLeft()) + " "
			} else {
				timeLeftBox = ""
			}

			if len(timeLeftBox) < b.opts.TimeBoxWidth {
				timeLeftBox = fmt.Sprintf("%s%s", strings.Repeat(" ", b.opts.TimeBoxWidth-len(timeLeftBox)), timeLeftBox)
			}
		}

		// speed
		if b.opts.ShowSpeed && b.units == united.UnitsBytes {
			if stats != nil {
				speedBox = stats.BPS().String() + " "
			} else {
				speedBox = ""
			}

			if len(speedBox) < b.opts.SpeedBoxWidth {
				speedBox = fmt.Sprintf("%s%s", strings.Repeat(" ", b.opts.SpeedBoxWidth-len(speedBox)), speedBox)
			}
		}
	}

	prefix := ""
	if b.prefix != "" {
		prefix = b.prefix + " "
	}

	postfix := ""
	if b.postfix != "" {
		postfix = " " + b.postfix
	}

	barWidth := escapeAwareRuneCountInString(countersBox + th.BarStart + th.BarEnd + percentBox + timeLeftBox + speedBox + prefix + postfix)

	// bar
	{
		fullSize := min(b.opts.BarWidth, width-barWidth)
		size := int(math.Ceil(float64(fullSize) * b.scale))
		padSize := fullSize - size
		if size > 0 {
			{
				curCount := int(math.Ceil(current * float64(size)))
				emptCount := size - curCount
				barBox = th.BarStart
				if emptCount < 0 {
					emptCount = 0
				}
				if curCount > size {
					curCount = size
				}
				barBox += strings.Repeat(th.Current, curCount)
				barBox += strings.Repeat(th.Empty, emptCount)
			}
			if padSize > 0 {
				barBox += strings.Repeat(" ", padSize-1)
			}
			barBox += th.BarEnd
		} else if padSize > 0 {
			barBox += th.BarStart + strings.Repeat(" ", padSize-1) + th.BarEnd
		}
	}

	// check len
	out = prefix + countersBox + barBox + percentBox + speedBox + timeLeftBox + postfix
	if escapeAwareRuneCountInString(out) < width {
		end = strings.Repeat(" ", width-utf8.RuneCountInString(out))
	}

	// and print!
	b.opts.Printf("%s", "\r"+out+end)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b *bar) update() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.write()
}

// Internal loop for writing progressbar
func (b *bar) writer() {
	b.update()
	for {
		select {
		case <-b.finishChan:
			return
		case <-time.After(b.opts.RefreshRate):
			b.update()
		}
	}
}
