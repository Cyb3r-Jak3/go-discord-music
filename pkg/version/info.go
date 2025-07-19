package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Version = "unknown"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	versionString := fmt.Sprintf("%s (Commit %s) (built %s)", Version, Commit, Date)
	if buildInfo, available := debug.ReadBuildInfo(); available {
		versionString = fmt.Sprintf("%s (Commit %s) (built %s with %s)", Version, Commit, Date, buildInfo.GoVersion)
	}
	return versionString
}
