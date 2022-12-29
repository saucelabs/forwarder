// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package pac

import "github.com/dop251/goja"

func (pr *ProxyResolver) TestingEval(script string) (goja.Value, error) {
	return pr.vm.RunString(script)
}
