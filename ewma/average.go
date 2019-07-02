package ewma

const (
	// averageMetricAge average over a 10-second period, which means the average
	// age of the metrics is in the period of 5 seconds
	averageMetricAge float64 = 5.0

	// decay formula for computing the decay factor for average metric age
	decay float64 = 2 / (float64(averageMetricAge) + 1)
)

// Average represents the exponentially weighted moving average of a series of numbers.
type Average interface {
	// Add a value to the series and update the moving average.
	Add(value float64)
	// Value returns the current value of the moving average.
	Value() float64
}

// New makes a new moving average, with an initial value
func New(initial float64) Average {
	return &average{
		value: initial,
	}
}

type average struct {
	value float64 // The current value of the average.
}

func (a *average) Add(value float64) {
	if a.value == 0 { // perhaps first input, no decay factor needed
		a.value = value
		return
	}
	a.value = (value * decay) + (a.value * (1 - decay))
}

func (a *average) Value() float64 {
	return a.value
}
