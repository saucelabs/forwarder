// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package pac

import _ "embed"

// asciiPacUtilsScript is a copy of the Mozilla PAC utils.js file.
//
//go:generate curl -sS https://raw.githubusercontent.com/mozilla/gecko-dev/master/netwerk/base/ascii_pac_utils.js -o ascii_pac_utils.js
//go:embed ascii_pac_utils.js
var asciiPacUtilsScript string
