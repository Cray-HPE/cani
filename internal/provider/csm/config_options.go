/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package csm

import (
	"context"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
)

func (csm *CSM) ConfigOptions(ctx context.Context) (provider.ConfigOptions, error) {
	providerConfig := provider.ConfigOptions{}

	// get valid roles and subroles from hsm
	values, _, err := csm.hsmClient.ServiceInfoApi.DoValuesGet(ctx)
	if err != nil {
		return providerConfig, err
	}
	providerConfig.ValidRoles = values.Role
	providerConfig.ValidSubRoles = values.SubRole

	// todo get these values from bss in the Global bootparameters in
	// the fields: kubernetes-pods-cidr and kubernetes-services-cidr
	providerConfig.K8sPodsCidr = "10.32.0.0/12"
	providerConfig.K8sServicesCidr = "10.16.0.0/12"

	tbv := &validate.ToBeValidated{}
	tbv.K8sPodsCidr = providerConfig.K8sPodsCidr
	tbv.K8sServicesCidr = providerConfig.K8sServicesCidr
	tbv.ValidRoles = providerConfig.ValidRoles
	tbv.ValidSubRoles = providerConfig.ValidSubRoles

	csm.TBV = tbv

	return providerConfig, nil
}
