// Copyright 2021 The customerror Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package customerror

import (
	"errors"
	"fmt"
)

// Demonstrates how to create static, and dynamic custom errors, also how to
// check, and instrospect custom errors.
func ExampleNew() {
	// Custom static error definition.
	var ErrMissingID = NewMissingError("id", "E1010", nil)

	// Some function, for demo purpose.
	var SomeFunc = func(id string) error {
		if id == "" {
			// Usage of the custom static error.
			return ErrMissingID
		}

		// Dynamic custom error.
		return NewFailedToError("write to disk", "E1523", nil)
	}

	// Case: Without `id`, retuns `ErrMissingID`.
	if err := SomeFunc(""); err != nil {
		fmt.Println(errors.Is(err, ErrMissingID)) // true

		var cE *CustomError
		if errors.As(err, &cE) {
			fmt.Println(cE.StatusCode) // 400
		}

		fmt.Println(err) // E1010: missing id (400 - Bad Request)
	}

	// Case: With `id`, retuns dynamic error.
	if err := SomeFunc("12345"); err != nil {
		var cE *CustomError
		if errors.As(err, &cE) {
			fmt.Println(cE.StatusCode) // 500
		}

		fmt.Println(err) // E1523: failed to write to disk (500 - Internal Server Error)
	}

	// output:
	// true
	// 400
	// E1010: missing id (400 - Bad Request)
	// 500
	// E1523: failed to write to disk (500 - Internal Server Error)
}
