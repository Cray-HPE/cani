package serve

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// NewCommand creates the "serve" command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the Web server.",
		Long:  `Run the Web server.`,
		RunE:  serve,
	}
	return cmd
}

type PageData struct {
	Inventory           devicetypes.Inventory
	DeviceTypes         map[string]devicetypes.DeviceType
	Device              *devicetypes.CaniDeviceType
	DeviceTypesTypeList []string //[]devicetypes.Type
}

type DetailPageData struct {
	Hardware devicetypes.CaniDeviceType
}

var sampleInventory devicetypes.Inventory

// initSampleInventory creates a sample inventory with two hardware items.
func initSampleInventory() {
	// read file, and unmarshal into inventory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("failed to get home directory: %v", err)
	}

	// Build the full path to the database file.
	dbPath := filepath.Join(home, ".cani", "canidb.json")

	content, err := os.ReadFile(dbPath)
	if err != nil {
		log.Printf("failed to read file %s: %v", dbPath, err)
	}

	if err := json.Unmarshal(content, &sampleInventory); err != nil {
		log.Printf("failed to unmarshal JSON from %s: %v", dbPath, err)
	}

}

var funcMap = template.FuncMap{
	"devicetypeTypeEquals": func(t devicetypes.Type, s string) bool {
		return string(t) == s
	},
}

var deviceTypesList = []string{
	string(devicetypes.System),
	string(devicetypes.Cabinet),
	string(devicetypes.CDU),
}

// func serve(cmd *cobra.Command, args []string) error {
// 	log.Printf("%+v", "Starting server")
// 	initSampleInventory()
// 	// Add a custom function to your template
// 	allDeviceTypes := devicetypes.All()
// 	devicetypesTypes := devicetypes.AllTypesString()

// 	// serve existing assets (logo, css) from the docs directory
// 	fs := http.FileServer(http.Dir("docs"))
// 	http.Handle("/docs/", http.StripPrefix("/docs/", fs))

// 	// Dashboard
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		tpl := template.Must(
// 			template.ParseFiles(
// 				"templates/base.html",
// 				"templates/dashboard.html",
// 			),
// 		)
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", PageData{Inventory: sampleInventory, DeviceTypesTypeList: devicetypesTypes})
// 	})

// 	// Dashboard
// 	http.HandleFunc("/racks/", func(w http.ResponseWriter, r *http.Request) {
// 		tpl := template.Must(
// 			template.ParseFiles(
// 				"templates/base.html",
// 				"templates/racks.html",
// 			),
// 		)
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", PageData{Inventory: sampleInventory, DeviceTypesTypeList: devicetypesTypes})
// 	})

// 	// Dashboard
// 	http.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
// 		tpl := template.Must(
// 			template.ParseFiles(
// 				"templates/base.html",
// 				"templates/devices.html",
// 			),
// 		)
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", PageData{Inventory: sampleInventory, DeviceTypesTypeList: devicetypesTypes})
// 	})
// 	// Individual device page
// 	http.HandleFunc("/device/", func(w http.ResponseWriter, r *http.Request) {
// 		idStr := strings.TrimPrefix(r.URL.Path, "/device/")
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		dt, ok := sampleInventory.Devices[id]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		tpl := template.New("base").Funcs(funcMap)
// 		tpl = template.Must(tpl.ParseFiles(
// 			"templates/base.html",
// 			"templates/device.html",
// 		))

// 		data := PageData{
// 			Inventory:           sampleInventory,
// 			DeviceTypes:         allDeviceTypes,
// 			Device:              dt,
// 			DeviceTypesTypeList: devicetypesTypes,
// 		}

// 		// No need to add Funcs here as we already did it above
// 		tpl.ExecuteTemplate(w, "base", data)
// 	})

// 	http.HandleFunc("/device/update/", func(w http.ResponseWriter, r *http.Request) {
// 		// Only allow POST
// 		if r.Method != "POST" {
// 			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		// Extract the device ID from the URL.
// 		idStr := strings.TrimPrefix(r.URL.Path, "/device/update/")
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		if err := r.ParseForm(); err != nil {
// 			http.Error(w, "Invalid form", http.StatusBadRequest)
// 			return
// 		}

// 		// Retrieve the existing device
// 		hw, ok := sampleInventory.Devices[id]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		// Update fields from form values.
// 		hw.Name = r.FormValue("name")
// 		hw.Type = devicetypes.Type(r.FormValue("type"))
// 		hw.Status = r.FormValue("status")
// 		hw.Vendor = r.FormValue("vendor")
// 		hw.Model = r.FormValue("model")
// 		// For Parent, if provided as a valid UUID string:
// 		if parentStr := r.FormValue("parent"); parentStr != "" {
// 			if parentID, err := uuid.Parse(parentStr); err == nil {
// 				hw.Parent = parentID
// 			}
// 		}

// 		// Save the updated device back to the inventory.
// 		sampleInventory.Devices[id] = hw

// 		// Redirect to the device detail page.
// 		http.Redirect(w, r, "/device/"+idStr, http.StatusSeeOther)
// 	})

// 	// Device Types
// 	http.HandleFunc("/devicetypes/", func(w http.ResponseWriter, r *http.Request) {
// 		tpl := template.Must(
// 			template.ParseFiles(
// 				"templates/base.html",
// 				"templates/devicetypes.html",
// 			),
// 		)
// 		tpl.Funcs(funcMap) // Add the custom function map to the template
// 		tpl.ExecuteTemplate(w, "base", PageData{
// 			Inventory:           sampleInventory,
// 			DeviceTypes:         allDeviceTypes,
// 			DeviceTypesTypeList: devicetypesTypes,
// 		})
// 	})

// 	// Individual device type page
// 	http.HandleFunc("/devicetype/", func(w http.ResponseWriter, r *http.Request) {
// 		// extract the slug, e.g. /devicetype/my-device-type
// 		slug := strings.TrimPrefix(r.URL.Path, "/devicetype/")
// 		dt, ok := allDeviceTypes[slug]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		tpl := template.Must(template.ParseFiles(
// 			"templates/base.html",
// 			"templates/devicetype.html",
// 		))
// 		data := struct {
// 			Inventory  devicetypes.Inventory
// 			DeviceType devicetypes.DeviceType
// 		}{
// 			Inventory:  sampleInventory,
// 			DeviceType: dt,
// 		}
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", data)
// 	})

// 	// New handler: Staged Devices page
// 	http.HandleFunc("/staged", func(w http.ResponseWriter, r *http.Request) {
// 		tpl := template.Must(template.ParseFiles(
// 			"templates/base.html",
// 			"templates/staged.html",
// 		))
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", PageData{Inventory: sampleInventory, DeviceTypesTypeList: devicetypesTypes})
// 	})

// 	// Production update handler: moves a device to production
// 	http.HandleFunc("/device/production/", func(w http.ResponseWriter, r *http.Request) {
// 		idStr := strings.TrimPrefix(r.URL.Path, "/device/production/")
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		hw, ok := sampleInventory.Devices[id]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		// Update the device status
// 		hw.Status = "provisioned" //FIXME
// 		sampleInventory.Devices[id] = hw

// 		// Render a confirmation page for production update
// 		tpl := template.Must(template.ParseFiles(
// 			"templates/base.html",
// 			"templates/production.html",
// 		))
// 		data := PageData{
// 			Inventory:           sampleInventory,
// 			DeviceTypes:         allDeviceTypes,
// 			Device:              hw,
// 			DeviceTypesTypeList: devicetypesTypes,
// 		}
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", data)
// 	})

// 	// Maintenance update handler: moves a device to maintenance
// 	http.HandleFunc("/device/maintenance/", func(w http.ResponseWriter, r *http.Request) {
// 		idStr := strings.TrimPrefix(r.URL.Path, "/device/maintenance/")
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		hw, ok := sampleInventory.Devices[id]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		// Update the device status
// 		hw.Status = "staged" //FIXME
// 		sampleInventory.Devices[id] = hw

// 		// Render a confirmation page for maintenance update
// 		tpl := template.Must(template.ParseFiles(
// 			"templates/base.html",
// 			"templates/maintenance.html",
// 		))
// 		data := PageData{
// 			Inventory:           sampleInventory,
// 			DeviceTypes:         allDeviceTypes,
// 			Device:              hw,
// 			DeviceTypesTypeList: devicetypesTypes,
// 		}
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", data)
// 	})

// 	// Maintenance update handler: decommissions a device
// 	http.HandleFunc("/device/decommission/", func(w http.ResponseWriter, r *http.Request) {
// 		idStr := strings.TrimPrefix(r.URL.Path, "/device/decommission/")
// 		id, err := uuid.Parse(idStr)
// 		if err != nil {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		hw, ok := sampleInventory.Devices[id]
// 		if !ok {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		// Update the device status
// 		hw.Status = "decommissioned" //FIXME
// 		sampleInventory.Devices[id] = hw

// 		// Render a confirmation page for maintenance update
// 		tpl := template.Must(template.ParseFiles(
// 			"templates/base.html",
// 			"templates/decommission.html",
// 		))
// 		data := PageData{
// 			Inventory:           sampleInventory,
// 			DeviceTypes:         allDeviceTypes,
// 			Device:              hw,
// 			DeviceTypesTypeList: devicetypesTypes,
// 		}
// 		tpl.Funcs(funcMap) // Add t
// 		tpl.ExecuteTemplate(w, "base", data)
// 	})

// 	log.Printf("Server started on http://localhost:8080")
// 	return http.ListenAndServe(":8080", nil)
// }

func serve(cmd *cobra.Command, args []string) error {
	log.Printf("Starting server")

	// Load configuration and data
	config := NewServerConfig()
	inventory, err := LoadInventory()
	if err != nil {
		return err
	}

	allDeviceTypes := devicetypes.All()
	deviceTypesList := devicetypes.AllTypesString()

	// Create handlers
	handlers := NewHandlers(inventory, allDeviceTypes, deviceTypesList)

	// Setup routes
	SetupRoutes(handlers)

	log.Printf("Server started on http://%s%s", config.Host, config.Port)
	return http.ListenAndServe(config.Port, nil)
}
