package filesvc

import (
	"os"
	"path/filepath"
	"regexp"
)

var storedIDRegex = regexp.MustCompile(`^[0-9a-f]{32}$`)

// RemoveStored deletes uploads from disk once their owning row is gone.
// Ids that do not look like generated ids are ignored, so a value coming from
// anywhere but randomID can never point outside the storage directory.
func (s *Service) RemoveStored(ids []string) {
	for _, id := range ids {
		if !storedIDRegex.MatchString(id) {
			continue
		}
		_ = os.Remove(filepath.Join(s.storagePath, filepath.Base(id)))
	}
}
