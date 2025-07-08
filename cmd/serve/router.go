package serve

import (
	"net/http"
)

func SetupRoutes(handlers *Handlers) {
	// Static files
	fs := http.FileServer(http.Dir("docs"))
	http.Handle("/docs/", http.StripPrefix("/docs/", fs))

	// Dashboard and main pages
	http.HandleFunc("/", handlers.Dashboard)
	http.HandleFunc("/racks/", handlers.Racks)
	// http.HandleFunc("/devices/", handlers.Devices)
	// http.HandleFunc("/staged", handlers.Staged)

	// Device operations
	http.HandleFunc("/device/", handlers.DeviceDetail)
	http.HandleFunc("/device/update/", handlers.DeviceUpdate)
	http.HandleFunc("/device/production/", handlers.DeviceProduction)
	http.HandleFunc("/device/maintenance/", handlers.DeviceMaintenance)
	http.HandleFunc("/device/decommission/", handlers.DeviceDecommission)

	// Device types
	http.HandleFunc("/devicetypes/", handlers.DeviceTypes)
	http.HandleFunc("/devicetype/", handlers.DeviceTypeDetail)
}
