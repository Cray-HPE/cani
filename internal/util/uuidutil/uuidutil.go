package uuidutil

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

func Join(ids []uuid.UUID, sep string, ignoreIDs ...uuid.UUID) string {
	// Build up ignore lookup map
	ignoreMap := map[uuid.UUID]bool{}
	for _, id := range ignoreIDs {
		ignoreMap[id] = true
	}

	// Build up string versions of the UUIDs
	idStrs := []string{}
	for _, id := range ids {
		if ignoreMap[id] {
			continue
		}

		idStrs = append(idStrs, id.String())
	}

	sort.Strings(idStrs)

	return strings.Join(idStrs, sep)
}
