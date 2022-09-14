package forwarder

import (
	"bytes"
	"encoding/gob"

	"github.com/saucelabs/customerror"
)

var ErrFailedToCopyOptions = customerror.NewFailedToError("deepCopy options")

// Copy from `source` to `target`.
//
// Basic deep copy implementation.
func deepCopy(source, target interface{}) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(source); err != nil {
		return customerror.Wrap(ErrFailedToCopyOptions, err)
	}

	if err := gob.NewDecoder(buf).Decode(target); err != nil {
		return customerror.Wrap(ErrFailedToCopyOptions, err)
	}

	return nil
}
