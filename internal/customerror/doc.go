// Copyright 2021 The customerror Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package customerror provides the base block to create custom errors, and
// some built-in custom errors. Custom errors standardizes errors across
// applications. It provides context - a `Message` to an optional `Err`.
// Additionally a `Code` - for example "E1010", and `StatusCode` can be
// provided.
//
// Static Errors:
//
// Custom static errors such as `ErrMissingID` can easily be created, and
// re-used. Just create that, for example, with the `NewMissingError` built-in.
//
// Dynamic Errors:
//
// Allows to create in-place custom errors.
//
// Examples:
//
// See `example_test.go` or the Example section of the GoDoc documention.
package customerror
