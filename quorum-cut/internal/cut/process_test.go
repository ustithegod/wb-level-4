package cut

import "testing"

func TestProcessInputsFields(t *testing.T) {
	t.Parallel()

	inputs := []InputLine{
		{Seq: 0, Text: "a:b:c"},
		{Seq: 1, Text: "x:y:z"},
	}
	opts := Options{
		FieldList: "1,3",
		Delimiter: ":",
	}

	got, err := ProcessInputs(inputs, opts)
	if err != nil {
		t.Fatalf("ProcessInputs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(got))
	}
	if got[0].Text != "a:c" {
		t.Fatalf("first result = %q, want %q", got[0].Text, "a:c")
	}
	if got[1].Text != "x:z" {
		t.Fatalf("second result = %q, want %q", got[1].Text, "x:z")
	}
}

func TestProcessInputsSuppressNoDelimiter(t *testing.T) {
	t.Parallel()

	inputs := []InputLine{{Seq: 0, Text: "abc"}}
	opts := Options{
		FieldList:       "1",
		Delimiter:       ":",
		SuppressNoDelim: true,
	}

	got, err := ProcessInputs(inputs, opts)
	if err != nil {
		t.Fatalf("ProcessInputs() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(got))
	}
}
