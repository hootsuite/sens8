package output

import "testing"

func TestRending(t *testing.T) {
	m := &Metric{[]*Point{}}
	m.AddPoint(&Point{"foo", 0.1, 12})
	m.AddPoint(&Point{"bar", 1.0, 42})

	expectedOutput := "foo 0.100000 12\nbar 1.000000 42"

	if m.Render() != expectedOutput {
		t.Errorf("Wrong output: %s and it's %s", m.Render(), expectedOutput)
	}
}
