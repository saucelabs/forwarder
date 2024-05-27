package martian

import (
	"testing"
)

type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Mock Flusher to capture flush calls.
type mockFlusher struct {
	flushes int
}

func (m *mockFlusher) Flush() error {
	m.flushes++
	return nil
}

func TestFlushWriter(t *testing.T) {
	tests := []struct {
		name      string
		pattern   [2]byte
		input     []string
		wantFlush int
	}{
		{"no pattern", [2]byte{'a', 'b'}, []string{"hello world"}, 0},
		{"pattern", [2]byte{'a', 'b'}, []string{"ab"}, 1},
		{"pattern at start", [2]byte{'a', 'b'}, []string{"abhello world"}, 1},
		{"pattern at end", [2]byte{'a', 'b'}, []string{"hello worldab"}, 1},
		{"double write", [2]byte{'a', 'b'}, []string{"hello a", "bworld"}, 1},
		{"triple write", [2]byte{'a', 'b'}, []string{"hello ab", "world a", "b"}, 2},
		{"repeat last byte", [2]byte{'a', 'b'}, []string{"a", "b", "b"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &mockWriter{}
			f := &mockFlusher{}
			fw := newPatternFlushWriter(w, f, tt.pattern)

			for _, input := range tt.input {
				if _, err := fw.Write([]byte(input)); err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
			}

			if f.flushes != tt.wantFlush {
				t.Errorf("expected %d flushes, got %d", tt.wantFlush, f.flushes)
			}
		})
	}
}
