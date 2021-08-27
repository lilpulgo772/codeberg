// Copyright 2021 Codeberg e.V.. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// This file contains codeberg-specific additions to Gitea
// that can live outside of the upstream source code.

package repo

import (
	"strings"
)

// check if a repo topic is used as a Codeberg flag
func (t *Topic) IsCodebergFlag() bool {
	return strings.HasPrefix(t.Name, "cbf-")
}
