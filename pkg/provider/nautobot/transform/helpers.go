package transform

import (
	"encoding/json"
	"reflect"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// RefUUID extracts a UUID from any generated per-field reference union type
// (e.g. Device_Status_Id, Cable_Status_Id). Nautobot 3.2 emits a distinct
// oneOf(UUID|int) union per reference field; every such type implements
// json.Marshaler with the UUID as member 0. Marshaling then parsing keeps the
// read path type-agnostic. Returns uuid.Nil when the union is nil, holds an
// integer id, or is otherwise not a UUID.
func RefUUID(m json.Marshaler) uuid.UUID {
	if m == nil {
		return uuid.Nil
	}
	// Guard against a typed-nil pointer (e.g. a nil *Device_Status_Id) whose
	// value-receiver MarshalJSON would panic on dereference.
	if rv := reflect.ValueOf(m); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return uuid.Nil
	}
	b, err := m.MarshalJSON()
	if err != nil || len(b) == 0 {
		return uuid.Nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// SetRefUUID sets any generated per-field reference union type to a UUID by
// unmarshaling the JSON-encoded UUID into the union's member 0. The concrete
// per-field union type is named once at the call site's declaration.
func SetRefUUID(u json.Unmarshaler, id uuid.UUID) error {
	b, err := json.Marshal(id)
	if err != nil {
		return err
	}
	return u.UnmarshalJSON(b)
}

// refFields extracts the Id (as json.Marshaler) and Url (*string) from a
// Nautobot 3.2 reference value. References are now per-field inline structs
// shaped `struct{ Id *<Type>_<Field>_Id; ObjectType *string; Url *string }`,
// so a single named type can no longer cover them; reflection keeps access
// type-agnostic. ref may be a struct, a pointer to one, or nil.
func refFields(ref any) (json.Marshaler, *string) {
	if ref == nil {
		return nil, nil
	}
	v := reflect.ValueOf(ref)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, nil
	}
	var m json.Marshaler
	var url *string
	if f := v.FieldByName("Id"); f.IsValid() && f.CanInterface() {
		if mm, ok := f.Interface().(json.Marshaler); ok {
			m = mm
		}
	}
	if f := v.FieldByName("Url"); f.IsValid() && f.CanInterface() {
		if uu, ok := f.Interface().(*string); ok {
			url = uu
		}
	}
	return m, url
}

// refID extracts the UUID from a Nautobot reference value's Id union.
func refID(ref any) uuid.UUID {
	m, _ := refFields(ref)
	return RefUUID(m)
}

// refIDVal is retained for call-site compatibility; identical to refID.
func refIDVal(ref any) uuid.UUID {
	return refID(ref)
}

// tenantRefID is retained for call-site compatibility; identical to refID.
func tenantRefID(ref any) uuid.UUID {
	return refID(ref)
}

// resolveTenantRefName is retained for call-site compatibility; identical to
// resolveRefName.
func resolveTenantRefName(ref any, nameMap map[uuid.UUID]string) string {
	return resolveRefName(ref, nameMap)
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

// refDisplay returns the reference URL as a display fallback.
func refDisplay(ref any) string {
	_, url := refFields(ref)
	return strVal(url)
}

// emailVal safely dereferences a *openapi_types.Email to a string.
func emailVal(e *openapi_types.Email) string {
	if e == nil {
		return ""
	}
	return string(*e)
}

// convCustomFields converts the Nautobot 3.2 custom-fields map (whose values
// are pointers) into a flat map[string]any for CANI types.
func convCustomFields(cf *map[string]*interface{}) map[string]any {
	if cf == nil {
		return nil
	}
	out := make(map[string]any, len(*cf))
	for k, v := range *cf {
		if v != nil {
			out[k] = *v
		} else {
			out[k] = nil
		}
	}
	return out
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

// resolveRefName looks up the name for a Nautobot reference using the provided
// UUID-to-name map, falling back to the reference URL when the UUID is absent.
func resolveRefName(ref any, nameMap map[uuid.UUID]string) string {
	m, url := refFields(ref)
	if id := RefUUID(m); id != uuid.Nil {
		if name, ok := nameMap[id]; ok {
			return name
		}
	}
	return strVal(url)
}
