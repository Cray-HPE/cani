// MIT License
//
// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package sls

import (
	"context"
	"fmt"
	"sort"
	"sync"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/antihax/optional"
	"github.com/rs/zerolog/log"
)

func NewHardware(xname xnames.Xname, class sls_client.HardwareClass, extraProperties interface{}) sls_client.Hardware {
	return sls_client.Hardware{
		Xname:           xname.String(),
		Class:           class,
		ExtraProperties: extraProperties,

		// Calculate derived fields
		Parent:     xnametypes.GetHMSCompParent(xname.String()),
		TypeString: xname.Type(),
		Type:       sls_client.HardwareType(sls_common.HMSTypeToHMSStringType(xname.Type())), // The main lookup table is in the SLS package, TODO should maybe move that into this package
	}
}

func NewHardwarePostOpts(hardware sls_client.Hardware) *sls_client.HardwareApiHardwarePostOpts {
	return &sls_client.HardwareApiHardwarePostOpts{
		Body: optional.NewInterface(sls_client.HardwarePost{
			Xname:           hardware.Xname,
			Class:           &hardware.Class,
			ExtraProperties: &hardware.ExtraProperties,
		}),
	}
}

func NewHardwareXnamePutOpts(hardware sls_client.Hardware) *sls_client.HardwareApiHardwareXnamePutOpts {
	return &sls_client.HardwareApiHardwareXnamePutOpts{
		Body: optional.NewInterface(sls_client.HardwarePut{
			Class:           &hardware.Class,
			ExtraProperties: &hardware.ExtraProperties,
		}),
	}
}

func SortHardware(hardware []sls_client.Hardware) {
	sort.Slice(hardware, func(i, j int) bool {
		return hardware[i].Xname < hardware[j].Xname
	})
}

func SortHardwareReverse(hardware []sls_client.Hardware) {
	sort.Slice(hardware, func(i, j int) bool {
		return hardware[i].Xname > hardware[j].Xname
	})
}

// FilterHardware will apply the given filter to a map of generic hardware
func FilterHardware(allHardware map[string]sls_client.Hardware, filter func(sls_client.Hardware) (bool, error)) (map[string]sls_client.Hardware, error) {
	result := map[string]sls_client.Hardware{}

	for xname, hardware := range allHardware {
		ok, err := filter(hardware)
		if err != nil {
			return nil, err
		}

		if ok {
			result[xname] = hardware
		}
	}

	return result, nil
}

func FilterHardwareByType(allHardware map[string]sls_client.Hardware, types ...xnametypes.HMSType) (map[string]sls_client.Hardware, error) {
	return FilterHardware(allHardware, func(hardware sls_client.Hardware) (bool, error) {
		for _, hmsType := range types {
			if hardware.TypeString == hmsType {
				return true, nil
			}
		}
		return false, nil
	})
}

func DecodeExtraProperties[T any](hardware sls_client.Hardware) (*T, error) {
	epRaw, err := hardware.DecodeExtraProperties()
	if err != nil {
		return nil, err
	}

	if epRaw == nil {
		return nil, nil
	}

	ep, ok := epRaw.(T)
	if !ok {
		var expectedType T
		return nil, fmt.Errorf("unexpected provider metadata type (%T) expected (%T)", epRaw, expectedType)
	}
	return &ep, nil
}

func HardwareUpdate(slsClient *sls_client.APIClient, ctx context.Context, hardwareToUpdate map[string]sls_client.Hardware, workers int) error {
	var wg sync.WaitGroup
	queue := make(chan sls_client.Hardware, 10)
	// TODO need to collect errors
	// errors :=
	updateWorker := func(id int) {
		defer wg.Done()

		log.Trace().Int("worker", id).Msgf("SLS HardwareUpdate: Starting worker")
		for hardware := range queue {
			log.Info().Int("worker", id).Msgf("SLS HardwareUpdate: Updating SLS hardware: %s", hardware.Xname)
			// Perform a PUT against SLS
			_, r, err := slsClient.HardwareApi.HardwareXnamePut(ctx, hardware.Xname, NewHardwareXnamePutOpts(hardware))
			if err != nil {
				// TODO need to collect errors
				// return errors.Join(
				// 	fmt.Errorf("failed to update hardware (%s) from SLS", hardare.Xname),
				// 	err,
				// )
				log.Error().Err(err).Msg("failed to update SLS")
				continue
			}
			log.Trace().Int("status", r.StatusCode).Msg("SLS HardwareUpdate: Updated hardware to SLS")
		}
		log.Trace().Int("worker", id).Msgf("SLS HardwareUpdate: Stopping worker")

	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go updateWorker(i)
	}

	for _, hardware := range hardwareToUpdate {
		log.Trace().Msgf("SLS HardwareUpdate: Adding %s to queue", hardware.Xname)
		queue <- hardware
	}
	log.Trace().Msgf("SLS HardwareUpdate: Queue is closed")
	close(queue)

	log.Trace().Msgf("SLS HardwareUpdate: Waiting for workers to complete")
	wg.Wait()

	return nil
}
