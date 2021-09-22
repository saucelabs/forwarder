// Copyright 2021 The customerror Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package customerror

import (
	"fmt"
	"net/http"
)

//////
// Built-in.
//////

// NewFailedToError is the building block for errors usually thrown when some
// action failed, e.g: "Failed to create host". Default status code is `500`.
//
// Note: Status code can be redefined, call `SetStatusCode`.
func NewFailedToError(message string, code string, err error) error {
	return New(fmt.Sprintf("failed to %s", message), code, http.StatusInternalServerError, err)
}

// NewInvalidError is the building block for errors usually thrown when
// something fail validation, e.g: "Invalid port". Default status code is `400`.
//
// Note: Status code can be redefined, call `SetStatusCode`.
func NewInvalidError(message string, code string, err error) error {
	return New(fmt.Sprintf("invalid %s", message), code, http.StatusBadRequest, err)
}

// NewMissingError is the building block for errors usually thrown when required
// information is missing, e.g: "Missing host". Default status code is `400`.
//
// Note: Status code can be redefined, call `SetStatusCode`.
func NewMissingError(message string, code string, err error) error {
	return New(fmt.Sprintf("missing %s", message), code, http.StatusBadRequest, err)
}

// NewRequiredError is the building block for errors usually thrown when
// required information is missing, e.g: "Port is required". Default status code
// is `400`.
//
// Note: Status code can be redefined, call `SetStatusCode`.
func NewRequiredError(message string, code string, err error) error {
	return New(fmt.Sprintf("%s required", message), code, http.StatusBadRequest, err)
}
