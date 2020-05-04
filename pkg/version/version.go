package version

// the way via debug.BuildInfo does not work -- it'll always return "devel"
// thus the oldschool way: hardcoded

const version = "0.2.0-devel"

// Current feed2imap version
func Version() string {
	return version
}
