package output

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	wrongTimestampFormatingError = errors.New("Wrong timestamp field format")
	wrongValueFormatingError     = errors.New("Wrong value field format")
	wrongLineFormatingError      = errors.New("Wrong number of item per line")
)

type Point struct {
	Name      string
	Value     float64
	Timestamp int64
}

type Metric struct {
	Points []*Point
}

func (p Point) Render() string {
	return fmt.Sprintf("%s %f %d", p.Name, p.Value, p.Timestamp)
}

func (m *Metric) AddPoint(p *Point) {
	if p.Timestamp == 0 {
		p.Timestamp = time.Now().Unix()
	}

	m.Points = append(m.Points, p)
}

func (m *Metric) Render() string {
	output := []string{}

	for _, p := range m.Points {
		output = append(output, p.Render())
	}

	return strings.Join(output, "\n")
}

func ParseMetric(output string) (*Metric, error) {
	var (
		m     Metric
		lines = strings.Split(output, "\n")
	)

	for _, line := range lines {
		fields := strings.Split(line, " ")

		if len(fields) != 3 {
			return nil, wrongLineFormatingError
		}

		value, err := strconv.ParseFloat(fields[1], 64)

		if err != nil {
			return nil, wrongValueFormatingError
		}

		timestamp, err := strconv.ParseInt(fields[2], 10, 64)

		if err != nil {
			return nil, wrongTimestampFormatingError
		}

		m.AddPoint(&Point{fields[0], value, timestamp})
	}

	return &m, nil
}
