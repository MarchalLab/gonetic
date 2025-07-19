package types

import "testing"

func TestIsInteractionStringFormat(t *testing.T) {
	valid := "1;2;3"
	invalid := []string{
		"1-2",
		"aux1",
		"1;2",
		"1;2;3;4",
	}

	if !IsInteractionStringFormat(valid) {
		t.Errorf("expected valid format for string: %s", valid)
	}
	for _, testCase := range invalid {
		if IsInteractionStringFormat(testCase) {
			t.Errorf("expected invalid format for string: %s", invalid)
		}
	}
}
