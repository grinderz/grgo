package info

import "fmt"

var (
	CommitRefName  = "v0.0.0"                                                               //nolint:gochecknoglobals
	CommitShortSha = "00000000"                                                             //nolint:gochecknoglobals
	BuildTimestamp = "unknown"                                                              //nolint:gochecknoglobals
	Version        = fmt.Sprintf("%s-%s-%s", CommitRefName, CommitShortSha, BuildTimestamp) //nolint:gochecknoglobals
)
