package version

import (
	"fmt"
	"testing"
)

func TestParseValidAPIVersions(t *testing.T) {
	inputs := []string{
		"v1",
		"v1beta1",
		"v2alpha3",
		"v10gamma7",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseAPIVersion(input); err != nil {
				t.Fatalf("Should have parsed successfully, but did not: %v", err)
			}
		})
	}
}

func TestParseInvalidAPIVersions(t *testing.T) {
	inputs := []string{
		"",
		"test",
		"v2-bar",
		"v1.2",
		"1",
		"1beta1",
		"v1beta",
		"vbeta1",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if v, err := ParseAPIVersion(input); err == nil {
				t.Fatalf("Should have not parsed successfully, but did: %+v", v)
			}
		})
	}
}

func TestComparingAPIVersions(t *testing.T) {
	type comparison struct {
		left  string
		right string
	}

	testcases := []comparison{
		{"v1beta1", "v1"},
		{"v1alpha1", "v1"},
		{"v1alpha1", "v1beta1"},
		{"v1beta3", "v1beta10"},
		{"v1", "v2"},
		{"v1", "v2alpha1"},
		{"v1", "v2"},
		{"v3", "v10"},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("%v", testcase), func(t *testing.T) {
			leftV, err := ParseAPIVersion(testcase.left)
			if err != nil {
				t.Fatalf("Left version should have parsed successfully, but did: %v", err)
			}

			rightV, err := ParseAPIVersion(testcase.right)
			if err != nil {
				t.Fatalf("Right version should have parsed successfully, but did: %v", err)
			}

			if !leftV.LessThan(rightV) {
				t.Fatalf("Expected %v < %v, but did not get this result", leftV, rightV)
			}
		})
	}
}

func TestPreferredAPIVersion(t *testing.T) {
	type comparison struct {
		versions  []string
		preferred string
	}

	testcases := []comparison{
		{versions: []string{"v1"}, preferred: "v1"},
		{versions: []string{"v1beta1", "v1"}, preferred: "v1"},
		{versions: []string{"v1beta1", "v1alpha1"}, preferred: "v1beta1"},
		{versions: []string{"v1beta3", "v1beta10"}, preferred: "v1beta10"},
		{versions: []string{"v1", "v2"}, preferred: "v2"},
		{versions: []string{"v2", "v1"}, preferred: "v2"},
		{versions: []string{"v2beta1", "v1"}, preferred: "v1"},
		{versions: []string{"v2beta1", "v1alpha3"}, preferred: "v2beta1"},
		{versions: []string{"v2beta3", "v1gamma1"}, preferred: "v1gamma1"},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("%v", testcase), func(t *testing.T) {
			preferred, err := PreferredAPIVersion(testcase.versions)
			if err != nil {
				t.Fatalf("Failed to parse testcase data: %v", err)
			}

			if preferred.String() != testcase.preferred {
				t.Fatalf("Expected latest version to be %v, but got %v", testcase.preferred, preferred.String())
			}
		})
	}
}
