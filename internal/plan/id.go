package plan

import (
	"crypto/sha256"
	"encoding/hex"
)

// projectID keys a workspace to a stable directory name. It is a digest of the
// absolute path rather than the path itself, so a project directory name cannot
// escape into the storage layout or collide with the tree above it. The readable
// path is kept inside the plan, so a listing is never opaque.
func projectID(abs string) string {
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:])[:12]
}
