// Copyright 2021 The customerror Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package customerror

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// CustomError is the base block to create custom errors. It provides context -
// a `Message` to an optional `Err`. Additionally a `Code` - for example "E1010",
// and `StatusCode` can be provided.
type CustomError struct {
	// Code can be any custom code, e.g.: E1010.
	Code string `json:"code"`

	// Err optionally wraps the original error.
	Err error `json:"-"`

	// Human readable message. Minimum length: 3.
	Message string `json:"message" validate:"required,gte=3"`

	// StatusCode is a valid HTTP status code, e.g.: 404.
	StatusCode int `json:"-"`
}

//////
// Error interface implementation.
//////

// SetStatusCode sets the status code.
//
// Note: Calling this on a static error is dangerous as it will change the
// status code of all its references!
func (cE *CustomError) SetStatusCode(code int) *CustomError {
	cE.StatusCode = code

	return cE
}

// Error interface implementation returns the properly formatted error message.
func (cE *CustomError) Error() string {
	errMsg := cE.Message

	if cE.Code != "" {
		errMsg = fmt.Sprintf("%s: %s", cE.Code, errMsg)
	}

	if cE.StatusCode != 0 {
		errMsg = fmt.Sprintf("%s (%d - %s)", errMsg, cE.StatusCode, http.StatusText(cE.StatusCode))
	}

	if cE.Err != nil {
		errMsg = fmt.Errorf("%s. Original Error: %s", errMsg, cE.Err).Error()
	}

	return errMsg
}

// Unwrap interface implementation returns inner error.
func (err *CustomError) Unwrap() error {
	return err.Err
}

// Wrap `customError` around `err`.
func Wrap(customError, err error) error {
	return fmt.Errorf("%w. Wrapped Error: %s", customError, err)
}

//////
// Factory.
//////

// New creates custom errors. `message` is required. Failing to satisfy that
// will throw a fatal error.
func New(message, code string, statusCode int, err error) *CustomError {
	cE := &CustomError{
		Message: message,
	}

	if err := validator.New().Struct(cE); err != nil {
		log.Fatalf("Invalid custom error. %s\n", err)

		return nil
	}

	// Status code validation. Could be validated, with `validator`, but doing
	// that would make it required.
	if statusCode != 0 &&
		(statusCode < http.StatusContinue ||
			statusCode > http.StatusNetworkAuthenticationRequired) {
		log.Fatalf("Invalid custom error. Invalid status code: %d\n", statusCode)
	}

	cE.Code = code
	cE.Err = err
	cE.StatusCode = statusCode

	return cE
}
