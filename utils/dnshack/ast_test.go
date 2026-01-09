// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build dnshack

package dnshack

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func parseFile(t *testing.T, path string) (*token.FileSet, *ast.File) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	fset := token.NewFileSet()
	name := filepath.Base(path)
	f, err := parser.ParseFile(fset, name, content, parser.ParseComments)
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	return fset, f
}

func typeDeclarationBytes(t *testing.T, fset *token.FileSet, f ast.Node, name string) []byte {
	t.Helper()

	var buf bytes.Buffer
	ast.Inspect(f, func(n ast.Node) bool {
		if spec, ok := n.(*ast.TypeSpec); ok {
			if spec.Name.String() == name {
				conf := &printer.Config{Mode: printer.UseSpaces}
				if err := conf.Fprint(&buf, fset, n); err != nil {
					t.Fatalf("Error printing type declaration: %v", err)
				}
			}
		}
		return true
	})

	return buf.Bytes()
}
