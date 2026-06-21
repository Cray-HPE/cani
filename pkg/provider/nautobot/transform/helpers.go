package transform

import (
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// refID extracts a UUID from a BulkWritableCableRequestStatus reference.
func refID(ref *nautobotapi.BulkWritableCableRequestStatus) uuid.UUID {
	if ref == nil || ref.Id == nil {
		return uuid.Nil
	}
	u, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		return uuid.Nil
	}
	return uuid.UUID(u)
}

// refIDVal extracts a UUID from a non-pointer BulkWritableCableRequestStatus.
func refIDVal(ref nautobotapi.BulkWritableCableRequestStatus) uuid.UUID {
	return refID(&ref)
}

// tenantRefID extracts a UUID from a BulkWritableCircuitRequestTenant reference.
func tenantRefID(ref *nautobotapi.BulkWritableCircuitRequestTenant) uuid.UUID {
	if ref == nil || ref.Id == nil {
		return uuid.Nil
	}
	u, err := ref.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		return uuid.Nil
	}
	return uuid.UUID(u)
}

// resolveTenantRefName looks up the name for a tenant-style reference by UUID.
// It falls back to the reference URL when the target object was not fetched.
func resolveTenantRefName(ref *nautobotapi.BulkWritableCircuitRequestTenant, nameMap map[uuid.UUID]string) string {
	id := tenantRefID(ref)
	if id != uuid.Nil {
		if name, ok := nameMap[id]; ok {
			return name
		}
	}
	if ref == nil {
		return ""
	}
	return strVal(ref.Url)
}

// directUUID converts an openapi_types.UUID pointer to uuid.UUID.
func directUUID(id *openapi_types.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return uuid.UUID(*id)
}

// strVal safely dereferences a *string.
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// intVal safely dereferences a *int, returning 0 if nil.
func intVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// refDisplay returns the Display field from a BulkWritableCableRequestStatus.
func refDisplay(ref *nautobotapi.BulkWritableCableRequestStatus) string {
	if ref == nil {
		return ""
	}
	return strVal(ref.Url) // fallback - display not always available
}

// BuildStatusNameMap creates a lookup from status UUID to name.
func BuildStatusNameMap(statuses []nautobotapi.Status) map[uuid.UUID]string {
	m := make(map[uuid.UUID]string, len(statuses))
	for _, s := range statuses {
		if s.Id != nil {
			m[uuid.UUID(*s.Id)] = s.Name
		}
	}
	return m
}

// BuildRoleNameMap creates a lookup from role UUID to name.
func BuildRoleNameMap(roles []nautobotapi.Role) map[uuid.UUID]string {
	m := make(map[uuid.UUID]string, len(roles))
	for _, r := range roles {
		if r.Id != nil {
			m[uuid.UUID(*r.Id)] = r.Name
		}
	}
	return m
}

// resolveRefName looks up the name for a BulkWritableCableRequestStatus
// reference using the provided UUID-to-name map. Falls back to the URL
// if the UUID is not found in the map.
func resolveRefName(ref nautobotapi.BulkWritableCableRequestStatus, nameMap map[uuid.UUID]string) string {
	id := refIDVal(ref)
	if id != uuid.Nil {
		if name, ok := nameMap[id]; ok {
			return name
		}
	}
	return strVal(ref.Url)
}
