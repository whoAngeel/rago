package eval

import (
	"fmt"
	"strings"
)

// CheckHost refuses a non-local Qdrant host unless allowRemote is set, because
// the eval creates and rewrites a collection.
func CheckHost(host string, allowRemote bool) error {
	h := strings.ToLower(strings.Trim(strings.TrimSpace(host), "[]"))
	switch h {
	case "localhost", "127.0.0.1", "::1":
		return nil
	}
	if allowRemote {
		return nil
	}
	return fmt.Errorf("refusing to run: QDRANT_HOST %q is not local (localhost, 127.0.0.1, ::1); this command creates and rewrites a collection. Pass --allow-remote only if you are sure", host)
}
