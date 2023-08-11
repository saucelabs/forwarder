package osdns

import (
	"go/build"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDNSConfigIntegrity(t *testing.T) {
	const typeName = "dnsConfig"

	fset, f := parseFile(t, "dnsconf.go")
	b1 := typeDeclarationBytes(t, fset, f, typeName)
	if len(b1) == 0 {
		t.Fatalf("dnsConfig not found in %s", "dnsconf.go")
	}

	path := filepath.Join(build.Default.GOROOT, "src", "net", "dnsconfig.go")
	fset2, f2 := parseFile(t, path)
	b2 := typeDeclarationBytes(t, fset2, f2, typeName)
	if len(b2) == 0 {
		t.Fatalf("dnsConfig not found in %s", path)
	}

	if diff := cmp.Diff(b1, b2); diff != "" {
		t.Fatalf("dns configs are not equal %s", diff)
	}
}
