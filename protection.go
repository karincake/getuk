package getuk

import (
	"regexp"
	"sync/atomic"
)

// SQL injection protection, enabled by default. When enabled:
//   - search keywords (gormhelper `searchColumns`) are bound as query
//     parameters instead of being interpolated into the SQL string
//   - Sort only accepts plain column identifiers (letters, digits, `_`, `.`)
//     and silently drops anything else
//
// Disabling restores the legacy behaviour, which is unsafe for user input.
var protectionDisabled atomic.Bool

// SetSQLInjectionProtection enables or disables SQL injection protection.
// Intended to be called once at startup.
func SetSQLInjectionProtection(enabled bool) {
	protectionDisabled.Store(!enabled)
}

// SQLInjectionProtection reports whether SQL injection protection is enabled.
func SQLInjectionProtection() bool {
	return !protectionDisabled.Load()
}

var safeIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// check if the string is a plain (optionally dotted) column identifier
func isSafeIdentifier(s string) bool {
	return safeIdentifier.MatchString(s)
}
