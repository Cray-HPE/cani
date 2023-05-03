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
