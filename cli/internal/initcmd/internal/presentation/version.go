package presentation

import "runtime/debug"

// cliVersion reports the tag the binary was stamped with; build-time Fang version logic reads the
// same build info, so it matches what a `--version` output should say.
func cliVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "(devel)"
	}

	return info.Main.Version
}
