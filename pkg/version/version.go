package version

// this is set by the linker during build
var (
	version = "0.5.0-post"
	commit  = ""
)

// Version returns the current feed2imap-go version
func Version() string {
	return version
}

// FullVersion returns the version including the commit hash
func FullVersion() string {
	return "Version " + version + " Commit: " + commit
}
