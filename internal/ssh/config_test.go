package ssh

import "testing"

func TestFirstOrEmpty(t *testing.T) {
	got := firstOrEmpty([]string{"a", "b"})
	if got != "a" {
		t.Fatalf("expected first value, got %q", got)
	}
}

func TestUniqueStrings(t *testing.T) {
	got := uniqueStrings([]string{"a", "a", "b", "", "c"})
	if len(got) != 3 {
		t.Fatalf("expected dedupe count 3, got %d", len(got))
	}
}
