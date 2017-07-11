package metric

type Point struct {
	// can be empty
	Suffix string
	Value  float64
}

type Metric interface {
	Collect() []Point
}
