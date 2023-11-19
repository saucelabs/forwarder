// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package header

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

type Action int

const (
	Remove Action = iota
	RemoveByPrefix
	Empty
	Add
)

var (
	DefaultAddRequestFunc  = map[string]func(*http.Request) string{}
	DefaultAddResponseFunc = map[string]func(*http.Response) string{}
)

type Header struct {
	Name   string
	Action Action
	Value  *string

	AddRequestFunc  map[string]func(*http.Request) string
	AddResponseFunc map[string]func(*http.Response) string
}

var (
	headerNameRegex = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	headerLineRegex = regexp.MustCompile(`^([A-Za-z0-9-]+):\s*(.*)\r?\n?$`)
)

// ParseHeader supports the following syntax:
// - "<name>: <value>" to add a header,
// - "<name>;" to set a header to empty,
// - "-<name>" to remove a header,
// - "-<name>*" to remove a header by prefix.
func ParseHeader(val string) (Header, error) {
	var h Header

	if strings.HasPrefix(val, "-") {
		if strings.HasSuffix(val, "*") {
			h.Name = val[1 : len(val)-1]
			h.Action = RemoveByPrefix
		} else {
			h.Name = val[1:]
			h.Action = Remove
		}
	} else {
		if strings.HasSuffix(val, ";") {
			h.Name = val[0 : len(val)-1]
			h.Action = Empty
		} else {
			if m := headerLineRegex.FindStringSubmatch(val); m != nil {
				h.Name = m[1]
				h.Value = &m[2]
				h.Action = Add
				h.AddRequestFunc = DefaultAddRequestFunc
				h.AddResponseFunc = DefaultAddResponseFunc
			} else {
				return Header{}, errors.New("invalid header value")
			}
		}
	}

	if !headerNameRegex.MatchString(h.Name) {
		return Header{}, errors.New("invalid header name")
	}

	return h, nil
}

func (h *Header) ApplyToRequest(req *http.Request) error {
	hh := req.Header
	switch h.Action {
	case Remove:
		hh.Del(h.Name)
	case RemoveByPrefix:
		removeHeadersByPrefix(hh, h.Name)
	case Empty:
		hh.Set(h.Name, "")
	case Add:
		if f, ok := isTemplateAddFunc(*h.Value); ok {
			if ff, ok := h.AddRequestFunc[f]; ok {
				hh.Set(h.Name, ff(req))
				return nil
			}
			return errors.New("add request func not found")
		}
		hh.Add(h.Name, *h.Value)
	}

	return nil
}

func (h *Header) ApplyToResponse(res *http.Response) error {
	hh := res.Header
	switch h.Action {
	case Remove:
		hh.Del(h.Name)
	case RemoveByPrefix:
		removeHeadersByPrefix(hh, h.Name)
	case Empty:
		hh.Set(h.Name, "")
	case Add:
		if f, ok := isTemplateAddFunc(*h.Value); ok {
			if ff, ok := h.AddResponseFunc[f]; ok {
				hh.Set(h.Name, ff(res))
				return nil
			}
			return errors.New("add response func not found")
		}
		hh.Add(h.Name, *h.Value)
	}

	return nil
}

func isTemplateAddFunc(s string) (string, bool) {
	if v, ok := strings.CutPrefix(s, "{{"); ok {
		if v, ok := strings.CutSuffix(v, "}}"); ok {
			return v, true
		}
	}
	return "", false
}

func removeHeadersByPrefix(h http.Header, prefix string) {
	for k := range h {
		if len(k) < len(prefix) {
			continue
		}
		if strings.EqualFold(k[0:len(prefix)], prefix) {
			h.Del(k)
		}
	}
}

func (h *Header) String() string {
	switch h.Action {
	case Remove:
		return "-" + h.Name
	case RemoveByPrefix:
		return "-" + h.Name + "*"
	case Empty:
		return h.Name + ";"
	case Add:
		return h.Name + ":" + *h.Value
	default:
		return ""
	}
}

type Headers []Header

func (s Headers) ModifyRequest(req *http.Request) error {
	for _, h := range s {
		if err := h.ApplyToRequest(req); err != nil {
			return err
		}
	}
	return nil
}

func (s Headers) ModifyResponse(res *http.Response) error {
	for _, h := range s {
		if err := h.ApplyToResponse(res); err != nil {
			return err
		}
	}
	return nil
}
