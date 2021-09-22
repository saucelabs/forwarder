// Copyright 2021 The randomness Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package randomness

import (
	"errors"
	"fmt"
	"log"
)

// Demonstrates how to generate a random number (integer)
func ExampleNew_generate() {
	r, err := New(0, 5, 0, false)
	if err != nil {
		log.Fatalln(err)
	}

	n, err := r.Generate()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(n < 0 && n > 5)

	// output:
	// false
}

// Demonstrates how to generate a random number (integer) - with the
// collision-free option, and no collision.
func ExampleNew_generate_collisionFree() {
	errMsgs := []error{}

	r, err := New(1, 10, 0, true)
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 3; i++ {
		_, err := r.Generate()

		if err != nil {
			errMsgs = append(errMsgs, err)
		}
	}

	saturated := false

	for _, err := range errMsgs {
		if errors.Is(err, ErrFailedToGenerateRangeSaturated) {
			saturated = true
		}
	}

	fmt.Println(saturated)

	// output:
	// false
}

// Demonstrates how to generate a random number (integer) - with the
// collision-free option, but causing collision.
func ExampleNew_generate_collisionFreeError() {
	errMsgs := []error{}

	r, err := New(1, 3, 0, true)
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 10; i++ {
		_, err := r.Generate()

		if err != nil {
			errMsgs = append(errMsgs, err)
		}
	}

	saturated := false

	for _, err := range errMsgs {
		if errors.Is(err, ErrFailedToGenerateRangeSaturated) {
			saturated = true
		}
	}

	fmt.Println(saturated)

	// output:
	// true
}

// Demonstrates how to generate a random number (integer) - with the
// collision-free option, no collision, and with maxRetry.
func ExampleNew_generate_collisionFreeMaxRetry() {
	errMsgs := []error{}

	r, err := New(1, 10, 100, true)
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 3; i++ {
		_, err := r.Generate()

		if err != nil {
			errMsgs = append(errMsgs, err)
		}
	}

	saturated := false
	reachedMaxRetries := false

	for _, err := range errMsgs {
		if errors.Is(err, ErrFailedToGenerateRangeSaturated) {
			saturated = true
		}

		if errors.Is(err, ErrFailedToGenerateReacedMaxRetry) {
			reachedMaxRetries = true
		}
	}

	fmt.Println(saturated)
	fmt.Println(reachedMaxRetries)

	// output:
	// false
	// false
}

// Demonstrates how to generate a random number (integer)
func ExampleNew_mustGenerate() {
	r, err := New(0, 5, 0, false)
	if err != nil {
		log.Fatalln(err)
	}

	n := r.MustGenerate()

	fmt.Println(n < 0 && n > 5)

	// output:
	// false
}
