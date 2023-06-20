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
package node

// import (
// 	"crypto/tls"
// 	"net/http"
// 	"testing"

// 	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
// 	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
// 	"github.com/hashicorp/go-retryablehttp"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/suite"
// )

// type AddNodeTestSuite struct {
// 	suite.Suite

// 	viper     *viper.Viper
// 	slsClient *sls_client.SLSClient
// 	hsmClient *hsm_client.HSMClient
// }

// func (suite *AddNodeTestSuite) SetupSuite() {
// 	// Setup viper
// 	suite.viper = viper.New()
// 	suite.viper.Set("hsm-url", "http://localhost:8080/apis/smd/hsm")
// 	suite.viper.Set("sls-url", "http://localhost:8080/apis/sls")
// 	suite.viper.Set("bss-url", "http://localhost:8080/apis/bss/boot")

// 	// Setup HTTP client
// 	httpClient := retryablehttp.NewClient()
// 	httpClient.HTTPClient.Transport = &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}

// 	// Create clients
// 	suite.slsClient = sls_client.NewSLSClient(suite.viper.GetString("sls-url"), httpClient.StandardClient(), "").WithAPIToken(token)
// 	suite.hsmClient = hsm_client.NewHSMClient(suite.viper.GetString("hsm-url"), httpClient.StandardClient(), "").WithAPIToken(token)

// }

// func (suite *AddNodeTestSuite) SetupTest() {
// 	// Reset SLS
// 	// Reset HSM
// 	// Reset BSS
// }

// func (suite *AddNodeTestSuite) TestAddRiverCompute_NewNode() {
// 	// Add a physical node

// 	// Verify services

// 	// Add a logical node

// 	// Verify services
// }

// func (suite *AddNodeTestSuite) TestAddRiverCompute_Replace() {
// 	// Add a logical node. This would be the case if the node was being replaced

// 	// Add a physical node
// }

// func TestAddNode(t *testing.T) {
// 	suite.Run(t, new(AddNodeTestSuite))
// }
