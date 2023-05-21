package main

import "testing"

func TestMultiValueFlagString(t *testing.T) {
	var xmvf = []struct {
		values multiValueFlag
		expect string
	}{
		{
			values: multiValueFlag{
				"test",
				"with multiples",
				"flags",
			},
			expect: `test, with multiples, flags`,
		},
		{
			values: multiValueFlag{
				"test",
			},
			expect: `test`,
		},
		{
			values: multiValueFlag{
				"",
			},
			expect: ``,
		},
		{
			values: nil,
			expect: ``,
		},
	}

	for _, mvf := range xmvf {
		got := mvf.values.String()
		if got != mvf.expect {
			t.Errorf("\nFor: %#v\nExpected: %v\nGot: %v", mvf.values, mvf.expect, got)
		}
	}
}
