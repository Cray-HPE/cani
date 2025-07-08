package serve

import (
	"html/template"
	_ "html/template"
	"net/http"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

type Handlers struct {
	inventory       *devicetypes.Inventory
	allDeviceTypes  map[string]devicetypes.DeviceType
	deviceTypesList []string
	templates       *TemplateManager
}

func NewHandlers(inventory *devicetypes.Inventory, allDeviceTypes map[string]devicetypes.DeviceType, deviceTypesList []string) *Handlers {
	return &Handlers{
		inventory:       inventory,
		allDeviceTypes:  allDeviceTypes,
		deviceTypesList: deviceTypesList,
		templates:       NewTemplateManager(),
	}
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Inventory:           *h.inventory,
		DeviceTypesTypeList: h.deviceTypesList,
	}
	h.templates.Render(w, "dashboard", data)
}

func (h *Handlers) DeviceDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/device/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	device, ok := h.inventory.Devices[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Inventory:           *h.inventory,
		DeviceTypes:         h.allDeviceTypes,
		Device:              device,
		DeviceTypesTypeList: h.deviceTypesList,
	}

	h.templates.Render(w, "device", data)
}

func (h *Handlers) DeviceUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/device/update/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	device, ok := h.inventory.Devices[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Update device fields
	h.updateDeviceFromForm(device, r)
	h.inventory.Devices[id] = device

	http.Redirect(w, r, "/device/"+idStr, http.StatusSeeOther)
}

func (h *Handlers) DeviceTypes(w http.ResponseWriter, r *http.Request) {
	// Device Types
	tpl := template.Must(
		template.ParseFiles(
			"templates/base.html",
			"templates/devicetypes.html",
		),
	)
	tpl.Funcs(funcMap) // Add the custom function map to the template
	tpl.ExecuteTemplate(w, "base", PageData{
		Inventory:           *h.inventory,
		DeviceTypes:         h.allDeviceTypes,
		DeviceTypesTypeList: h.deviceTypesList,
	})
}

func (h *Handlers) DeviceTypeDetail(w http.ResponseWriter, r *http.Request) {
	// Individual device type page
	// extract the slug, e.g. /devicetype/my-device-type
	slug := strings.TrimPrefix(r.URL.Path, "/devicetype/")
	dt, ok := h.allDeviceTypes[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}
	tpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/devicetype.html",
	))
	data := struct {
		Inventory  devicetypes.Inventory
		DeviceType devicetypes.DeviceType
	}{
		Inventory:  *h.inventory,
		DeviceType: dt,
	}
	tpl.Funcs(funcMap) // Add the custom function map to the template
	tpl.ExecuteTemplate(w, "base", data)
}

func (h *Handlers) DeviceMaintenance(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Inventory:           *h.inventory,
		DeviceTypesTypeList: h.deviceTypesList,
	}
	h.templates.Render(w, "base", data)
}

func (h *Handlers) DeviceProduction(w http.ResponseWriter, r *http.Request) {

	idStr := strings.TrimPrefix(r.URL.Path, "/device/production/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	device, ok := h.inventory.Devices[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.updateDeviceFromForm(device, r)
	h.inventory.Devices[id] = device

	http.Redirect(w, r, "/device/"+idStr, http.StatusSeeOther)
}

func (h *Handlers) DeviceDecommission(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/device/decommission/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	device, ok := h.inventory.Devices[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.updateDeviceFromForm(device, r)
	h.inventory.Devices[id] = device

	http.Redirect(w, r, "/device/"+idStr, http.StatusSeeOther)
}

func (h *Handlers) Racks(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Inventory:           *h.inventory,
		DeviceTypesTypeList: h.deviceTypesList,
	}
	h.templates.Render(w, "base", data)
}

func (h *Handlers) updateDeviceFromForm(device *devicetypes.CaniDeviceType, r *http.Request) {
	device.Name = r.FormValue("name")
	device.Type = devicetypes.Type(r.FormValue("type"))
	device.Status = r.FormValue("status")
	device.Vendor = r.FormValue("vendor")
	device.Model = r.FormValue("model")

	if parentStr := r.FormValue("parent"); parentStr != "" {
		if parentID, err := uuid.Parse(parentStr); err == nil {
			device.Parent = parentID
		}
	}
}
