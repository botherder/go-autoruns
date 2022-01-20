// This file is part of go-autoruns.
// Copyright (c) 2018-2021 Claudio Guarnieri
// See the file 'LICENSE' for copying permission.

package autoruns

import (
	"path"
	"strings"
)

func parseShellInvocation(shellLine string, autorun *Autorun) {
	autorun.LaunchString = strings.SplitAfter(shellLine, "=")[1]
	// We need to make sure to drop !! from paths
	autorun.ImagePath = strings.Replace(strings.Split(autorun.LaunchString, " ")[0], "!!", "", -1)
	autorun.ImageName = path.Base(autorun.ImagePath)

	args := strings.Split(autorun.LaunchString, " ")
	if len(args) > 1 {
		autorun.Arguments = strings.Join(args[1:], " ")
	}
}
