package utils

import "testing"

// ValidateStringParameter checks if the input strings are equal and
// adds a test error if they are not
func ValidateStringParameter(
	actual string,
	expected string,
	parameterName string,
	t *testing.T,
) {
	if actual != expected {
		t.Errorf(
			"Expected \"%s\" to be \"%s\" but got \"%s\" instead",
			parameterName,
			expected,
			actual,
		)
	}
}

// ValidateError checks if the input errors are equal and
// adds a test error if they are not
func ValidateError(actual error, expected error, t *testing.T) {
	actualIsNil := actual == nil
	expectedIsNil := expected == nil

	if actualIsNil && expectedIsNil {
		return
	}

	// XOR the hard way because Go doesn't have the logical XOR operator...
	if actualIsNil != expectedIsNil ||
		actual.Error() != expected.Error() {

		t.Errorf(
			"Expected error to be \"%v\" but got \"%v\" instead",
			expected,
			actual,
		)
	}
}
