package united

// Units is a enum that represents task progression units
type Units int

const (
	// UnitsNone represents unit-less values
	UnitsNone Units = iota
	// UnitsBytes if formatted as B, KiB, MiB, etc.
	UnitsBytes
)
