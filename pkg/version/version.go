package version

var (
	// Version is the version filled in by the compiler
	Version string
	// GitCommit is the git commit filled in by the compiler
	GitCommit string
)

// GetVersion returns the version of NDM
func GetVersion() string {
	return Version
}

// GetGitCommit returns commit from which this version is compiled
func GetGitCommit() string {
	return GitCommit
}
