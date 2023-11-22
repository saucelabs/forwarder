// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import _ "embed"

// asciiPacUtilsScript is a copy of the Mozilla PAC utils.js file.
//
//go:generate curl -sS https://raw.githubusercontent.com/mozilla/gecko-dev/master/netwerk/base/ascii_pac_utils.js -o ascii_pac_utils.js
//go:embed ascii_pac_utils.js
var asciiPacUtilsScript string
