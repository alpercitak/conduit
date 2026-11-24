package core

import "testing"

func TestTracerEvictsOldest(t *testing.T) {
	tr := NewInMemoryTracer()
	tr.limit = 2

	for _, id := range []string{"a", "b", "a", "c"} {
		tr.Mark(&Context{RequestID: id}, "x")
	}

	if _, ok := tr.Get("a"); ok {
		t.Fatal("want oldest trace evicted")
	}
	if ev, ok := tr.Get("b"); !ok || len(ev) != 1 {
		t.Fatalf("want b kept, got %v", ev)
	}
	if _, ok := tr.Get("c"); !ok {
		t.Fatal("want c kept")
	}
	if len(tr.traces) != 2 {
		t.Fatalf("want 2 traces, got %d", len(tr.traces))
	}
}
