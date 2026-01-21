package orale

// Loader is a struct that contains all the values loaded from flags, environment
// variables, and configuration files. It can be used to marshal the values into
// a struct.
type Loader struct {
	// FlagValues is a map of flag values by path.
	FlagValues map[string][]any
	// EnvironmentValues is a map of environment variable values by path.
	EnvironmentValues map[string][]any
	// ConfigurationFiles is a slice of configuration files.
	ConfigurationFiles []*File
	// expandTargets is used to detect circular variable expansion.
	expandTargets map[string]struct{}
	// expandCache caches expanded values for performance.
	expandCache map[string]string
	// DisableDecryptionWarnings suppresses warnings when encrypted values cannot
	// be decrypted. When decryption fails, the value will be replaced with an
	// empty string regardless of this setting.
	DisableDecryptionWarnings bool
}
