package buildinfo

import "fmt"

var (
	Version = "0.1.0-dev.0+unknown"
	Commit  = "unknown"
	Date    = "unknown"
	Dirty   = "false"
)

func String() string {
	return fmt.Sprintf("%s (commit %s, built %s, dirty %s)", Version, Commit, Date, Dirty)
}
