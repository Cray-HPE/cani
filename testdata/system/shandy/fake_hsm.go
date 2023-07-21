package main

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed hsm_state_components.json
var stateComponents []byte

//go:embed hsm_inventory_hardware.json
var inventoryHardware []byte

func main() {
	http.HandleFunc("/v2/State/Components", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Get /v2/State/Components")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(stateComponents)
	})
	http.HandleFunc("/v2/Inventory/Hardware", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Get /v2/Inventory/Hardware")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(inventoryHardware)
	})

	fmt.Println("Listening :8097")
	http.ListenAndServe(":8097", nil)
}
