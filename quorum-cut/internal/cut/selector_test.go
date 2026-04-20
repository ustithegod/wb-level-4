package cut

import "testing"

func TestParseSelectorContains(t *testing.T) {
	t.Parallel()

	selector, err := parseSelector("1,3-4,7-")
	if err != nil {
		t.Fatalf("parseSelector() error = %v", err)
	}

	cases := map[int]bool{
		1: true,
		2: false,
		3: true,
		4: true,
		5: false,
		7: true,
		8: true,
	}

	for idx, want := range cases {
		if got := selector.contains(idx); got != want {
			t.Fatalf("contains(%d) = %v, want %v", idx, got, want)
		}
	}
}
