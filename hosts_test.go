package forwarder

import "testing"

func TestReadHostsFile(t *testing.T) {
	m, err := ReadHostsFile()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		t.Logf("%s=%s", k, v)
	}
	if len(m) == 0 {
		t.Fatal("no hosts found")
	}
	if m["localhost"] == nil {
		t.Fatal("localhost not found")
	}
}
