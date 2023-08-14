package osdns

import (
	"go/build"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolverConfigIntegrity(t *testing.T) {
	const typeName = "resolverConfig"

	fset, f := parseFile(t, "resolvconf.go")
	b1 := typeDeclarationBytes(t, fset, f, typeName)
	if len(b1) == 0 {
		t.Fatalf("resolverConfig not found in %s", "resolvconf.go")
	}

	path := filepath.Join(build.Default.GOROOT, "src", "net", "dnsclient_unix.go")
	fset2, f2 := parseFile(t, path)
	b2 := typeDeclarationBytes(t, fset2, f2, typeName)
	if len(b2) == 0 {
		t.Fatalf("resolverConfig not found in %s", path)
	}

	if diff := cmp.Diff(b1, b2); diff != "" {
		t.Fatalf("resolver configs are not equal %s", diff)
	}
}
