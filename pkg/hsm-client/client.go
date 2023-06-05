/*
 * Hardware State Manager API
 *
 * The Hardware State Manager (HSM) inventories, monitors, and manages hardware, and tracks the logical and dynamic component states, such as roles, NIDs, and other basic metadata needed to provide most common administrative and operational functions. HSM is the single source of truth for the state of the system. It contains the component state and information on Redfish endpoints for communicating with components via Redfish. It also allows administrators to create partitions and groups for other uses. ## Resources ### /State/Components HMS components are created during inventory discovery and provide a higher-level representation of the component, including state, NID, role (i.e. compute/service), subtype, and so on. Unlike ComponentEndpoints, however, they are not strictly linked to the parent RedfishEndpoint, and are not automatically deleted when the RedfishEndpoints are (though they can be deleted via a separate call). This is because these components can also represent abstract components, such as removed components (e.g. which would remain, but have their states changed to \"Empty\" upon removal). ### /Defaults/NodeMaps  This resource allows a mapping file (NodeMaps) to be uploaded that maps node xnames to Node IDs, and optionally, to roles and subroles. These mappings are used when discovering nodes for the first time. These mappings should be uploaded prior to discovery and should contain mappings for each valid node xname in the system, whether populated or not. Nodemap is a JSON file that contains the xname of the node, node ID, and optionally role and subrole. Role can be Compute, Application, Storage, Management etc. The NodeMaps collection can be uploaded to HSM automatically at install time by specifying it as a JSON file. As a result, the endpoints are then automatically discovered by REDS, and inventory discovery is performed by HSM. The desired NID numbers will be set as soon as the nodes are created using the NodeMaps collection.  It is recommended that Nodemaps are uploaded at install time before discovery happens. If they are uploaded after discovery, then the node xnames need to be manually updated with the correct NIDs. You can update NIDs for individual components by using PATCH /State/Components/{xname}/NID.  ### /Inventory/Hardware  This resource shows the hardware inventory of the entire system and contains FRU information in location. All entries are displayed as a flat array. ### /Inventory/HardwareByFRU  Every component has FRU information. This resource shows the hardware inventory for all FRUs or for a specific FRU irrespective of the location. This information is constant regardless of where the hardware item is currently in the system. If a HWInventoryByLocation entry is currently populated with a piece of hardware, it will have the corresponding HWInventoryByFRU object embedded. This FRU info can also be looked up by FRU ID regardless of the current location. ### /Inventory/Hardware/Query/{xname}  This resource gets you information about a specific component and it's sub-components. The xname can be a component, partition, ALL, or s0. Both ALL and s0 represent the entire system. ### /Inventory/RedfishEndpoints  This is a BMC or other Redfish controller that has a Redfish entry point and Redfish service root. It is used to discover the components managed by this endpoint during discovery and handles all Redfish interactions by these subcomponents.  If the endpoint has been discovered, this entry will include the ComponentEndpoint entries for these managed subcomponents. You can also create a Redfish Endpoint or update the definition for a Redfish Endpoint. The xname identifies the location of all components in the system, including chassis, controllers, nodes, and so on. Redfish endpoints are given to State Manager. ### /Inventory/ComponentEndpoints  Component Endpoints are the specific URLs for each individual component that are under the Redfish endpoint. Component endpoints are discovered during inventory discovery. They are the management-plane representation of system components and are linked to the parent Redfish Endpoint. They provide a glue layer to bridge the higher-level representation of a component with how it is represented locally by Redfish.  The collection of ComponentEndpoints can be obtained in full, optionally filtered on certain criteria (e.g. obtain just Node components), or accessed by their xname IDs individually. ### /Inventory/ServiceEndpoints  ServiceEndpoints help you do things on Redfish like updating the firmware. They are discovered during inventory discovery. ### /groups  Groups are named sets of system components, most commonly nodes. A group groups components under an administratively chosen label (group name). Each component may belong to any number of groups. If a group has exclusiveGroup=<excl-label> set, then a node may only be a member of one group that matches that exclusive label. For example, if the exclusive group label 'colors' is associated with groups 'blue', 'red', and 'green', then a component that is part of 'green' could not also be placed in 'red'. You can create, modify, or delete a group and its members. You can also use group names as filters for API calls. ### /partitions  A partition is a formal, non-overlapping division of the system that forms an administratively distinct sub-system. Each component may belong to at most one partition. Partitions are used as an access control mechanism or for implementing multi-tenancy. You can create, modify, or delete a partition and its members. You can also use partitions as filters for other API calls. ### /memberships  A membership shows the association of a component xname to its set of group labels and partition names. There can be many group labels and up to one partition per component. Memberships are not modified directly, as the underlying group or partition is modified instead. A component can be removed from one of the listed groups or partitions or added via POST as well as being present in the initial set of members when a partition or group is created. You can retrieve the memberships for components or memberships for a specific xname. ### /Inventory/DiscoveryStatus  Check discovery status for all components or you can track the status for a specific job ID. You can also check per-endpoint discover status for each RedfishEndpoint. Contains status information about the discovery operation for clients to query. The discover operation returns a link or links to status objects so that a client can determine when the discovery operation is complete. ### /Inventory/Discover  Discover subcomponents by querying all RedfishEndpoints. Once the RedfishEndpoint objects are created, inventory discovery will query these controllers and create or update management plane and managed plane objects representing the components (e.g. nodes, node enclosures, node cards for Mountain chassis CMM endpoints). ### /Subscriptions/SCN  Manage subscriptions to state change notifications (SCNs) from HSM. You can also subscribe to state change notifications by using the HMS Notification Fanout Daemon API. ## Workflows  ### Add and Delete a Redfish Endpoint #### POST /Inventory/RedfishEndpoints When you manually create Redfish endpoints, the discovery is automatically initiated. You would create Redfish endpoints for components that are not automatically discovered by REDS or MEDS. #### GET /Inventory/RedfishEndpoints Check the Redfish endpoints that have been added and check the status of discovery. #### DELETE /Inventory/RedfishEndpoints/{xname} Delete a specific Redfish endpoint. ### Perform Inventory Discovery #### POST /Inventory/Discover Start inventory discovery of a system's subcomponents by querying all Redfish endpoints. If needed, specify an ID or hostname (xname) in the payload. #### GET /Inventory/DiscoveryStatus Check the discovery status of all Redfish endpoints. You can also check the discovery status for each individual component by providing ID. ### Query and Update HMS Components (State/NID) #### GET /State/Components Retrieve all HMS Components found by inventory discovery as a named (\"Components\") array.  #### PATCH /State/Components/{xname}/Enabled Modify the component's Enabled field.  #### DELETE /State/Components/{xname} Delete a specific HMS component by providing its xname. As noted, components are not automatically deleted when RedfishEndpoints or ComponentEndpoints are deleted. ### Create and Delete a New Group #### GET /hsm/v2/State/Components Retrieve a list of desired components and their state. Select the nodes that you want to group.  #### POST /groups Create the new group with desired members. Provide a group label (required), description, name, members etc. in the JSON payload. #### GET /groups/{group_label} Retrieve the group that was create with the label. #### GET /State/Components/{group_label} Retrieve the current state for all the components in the group. #### DELETE /groups/{group_label} Delete the group specified by {group_label}. ## Valid State Transitions ``` Prior State -> New State     - Reason Ready       -> Standby       - HBTD if node has many missed heartbeats Ready       -> Ready/Warning - HBTD if node has a few missed heartbeats Standby     -> Ready         - HBTD Node re-starts heartbeating On          -> Ready         - HBTD Node started heartbeating Off         -> Ready         - HBTD sees heartbeats before Redfish Event (On) Standby     -> On            - Redfish Event (On) or if re-discovered while in the standby state Off         -> On            - Redfish Event (On) Standby     -> Off           - Redfish Event (Off) Ready       -> Off           - Redfish Event (Off) On          -> Off           - Redfish Event (Off) Any State   -> Empty         - Redfish Endpoint is disabled meaning component removal ``` Generally, nodes transition 'Off' -> 'On' -> 'Ready' when going from 'Off' to booted, and 'Ready' -> 'Ready/Warning' -> 'Standby' -> 'Off' when shutdown.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package hsm_client

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

var (
	jsonCheck = regexp.MustCompile("(?i:[application|text]/json)")
	xmlCheck  = regexp.MustCompile("(?i:[application|text]/xml)")
)

// APIClient manages communication with the Hardware State Manager API API v1.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// API Services

	AdminLocksApi *AdminLocksApiService

	AdminReservationsApi *AdminReservationsApiService

	CliDangerThisWillDeleteAllComponentEndpointsContinueApi *CliDangerThisWillDeleteAllComponentEndpointsContinueApiService

	CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApi *CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApiService

	CliDangerThisWillDeleteAllComponentsInHSMContinueApi *CliDangerThisWillDeleteAllComponentsInHSMContinueApiService

	CliDangerThisWillDeleteAllFRUsForHSMContinueApi *CliDangerThisWillDeleteAllFRUsForHSMContinueApiService

	CliDangerThisWillDeleteAllHardwareHistoryContinueApi *CliDangerThisWillDeleteAllHardwareHistoryContinueApiService

	CliDangerThisWillDeleteAllHardwareInventoryContinueApi *CliDangerThisWillDeleteAllHardwareInventoryContinueApiService

	CliDangerThisWillDeleteAllHistoryForThisFRUContinueApi *CliDangerThisWillDeleteAllHistoryForThisFRUContinueApiService

	CliDangerThisWillDeleteAllHistoryForThisXnameContinueApi *CliDangerThisWillDeleteAllHistoryForThisXnameContinueApiService

	CliDangerThisWillDeleteAllNodeMapsContinueApi *CliDangerThisWillDeleteAllNodeMapsContinueApiService

	CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApi *CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApiService

	CliDangerThisWillDeleteAllServiceEndpointsContinueApi *CliDangerThisWillDeleteAllServiceEndpointsContinueApiService

	CliIgnoreApi *CliIgnoreApiService

	ComponentApi *ComponentApiService

	ComponentEndpointApi *ComponentEndpointApiService

	ComponentEthernetInterfacesApi *ComponentEthernetInterfacesApiService

	DiscoverApi *DiscoverApiService

	DiscoveryStatusApi *DiscoveryStatusApiService

	GroupApi *GroupApiService

	HWInventoryApi *HWInventoryApiService

	HWInventoryByFRUApi *HWInventoryByFRUApiService

	HWInventoryByLocationApi *HWInventoryByLocationApiService

	HWInventoryHistoryApi *HWInventoryHistoryApiService

	LockingApi *LockingApiService

	MembershipApi *MembershipApiService

	NodeMapApi *NodeMapApiService

	PartitionApi *PartitionApiService

	PowerMapApi *PowerMapApiService

	RedfishEndpointApi *RedfishEndpointApiService

	SCNApi *SCNApiService

	ServiceEndpointApi *ServiceEndpointApiService

	ServiceInfoApi *ServiceInfoApiService

	ServiceReservationsApi *ServiceReservationsApiService
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) *APIClient {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = retryablehttp.NewClient()
	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	c.AdminLocksApi = (*AdminLocksApiService)(&c.common)
	c.AdminReservationsApi = (*AdminReservationsApiService)(&c.common)
	c.CliDangerThisWillDeleteAllComponentEndpointsContinueApi = (*CliDangerThisWillDeleteAllComponentEndpointsContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApi = (*CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllComponentsInHSMContinueApi = (*CliDangerThisWillDeleteAllComponentsInHSMContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllFRUsForHSMContinueApi = (*CliDangerThisWillDeleteAllFRUsForHSMContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllHardwareHistoryContinueApi = (*CliDangerThisWillDeleteAllHardwareHistoryContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllHardwareInventoryContinueApi = (*CliDangerThisWillDeleteAllHardwareInventoryContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllHistoryForThisFRUContinueApi = (*CliDangerThisWillDeleteAllHistoryForThisFRUContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllHistoryForThisXnameContinueApi = (*CliDangerThisWillDeleteAllHistoryForThisXnameContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllNodeMapsContinueApi = (*CliDangerThisWillDeleteAllNodeMapsContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApi = (*CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApiService)(&c.common)
	c.CliDangerThisWillDeleteAllServiceEndpointsContinueApi = (*CliDangerThisWillDeleteAllServiceEndpointsContinueApiService)(&c.common)
	c.CliIgnoreApi = (*CliIgnoreApiService)(&c.common)
	c.ComponentApi = (*ComponentApiService)(&c.common)
	c.ComponentEndpointApi = (*ComponentEndpointApiService)(&c.common)
	c.ComponentEthernetInterfacesApi = (*ComponentEthernetInterfacesApiService)(&c.common)
	c.DiscoverApi = (*DiscoverApiService)(&c.common)
	c.DiscoveryStatusApi = (*DiscoveryStatusApiService)(&c.common)
	c.GroupApi = (*GroupApiService)(&c.common)
	c.HWInventoryApi = (*HWInventoryApiService)(&c.common)
	c.HWInventoryByFRUApi = (*HWInventoryByFRUApiService)(&c.common)
	c.HWInventoryByLocationApi = (*HWInventoryByLocationApiService)(&c.common)
	c.HWInventoryHistoryApi = (*HWInventoryHistoryApiService)(&c.common)
	c.LockingApi = (*LockingApiService)(&c.common)
	c.MembershipApi = (*MembershipApiService)(&c.common)
	c.NodeMapApi = (*NodeMapApiService)(&c.common)
	c.PartitionApi = (*PartitionApiService)(&c.common)
	c.PowerMapApi = (*PowerMapApiService)(&c.common)
	c.RedfishEndpointApi = (*RedfishEndpointApiService)(&c.common)
	c.SCNApi = (*SCNApiService)(&c.common)
	c.ServiceEndpointApi = (*ServiceEndpointApiService)(&c.common)
	c.ServiceInfoApi = (*ServiceInfoApiService)(&c.common)
	c.ServiceReservationsApi = (*ServiceReservationsApiService)(&c.common)

	return c
}

func atoi(in string) (int, error) {
	return strconv.Atoi(in)
}

// selectHeaderContentType select a content type from the available list.
func selectHeaderContentType(contentTypes []string) string {
	if len(contentTypes) == 0 {
		return ""
	}
	if contains(contentTypes, "application/json") {
		return "application/json"
	}
	return contentTypes[0] // use the first content type specified in 'consumes'
}

// selectHeaderAccept join all accept types and return
func selectHeaderAccept(accepts []string) string {
	if len(accepts) == 0 {
		return ""
	}

	if contains(accepts, "application/json") {
		return "application/json"
	}

	return strings.Join(accepts, ",")
}

// contains is a case insenstive match, finding needle in a haystack
func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if strings.ToLower(a) == strings.ToLower(needle) {
			return true
		}
	}
	return false
}

// Verify optional parameters are of the correct type.
func typeCheckParameter(obj interface{}, expected string, name string) error {
	// Make sure there is an object.
	if obj == nil {
		return nil
	}

	// Check the type is as expected.
	if reflect.TypeOf(obj).String() != expected {
		return fmt.Errorf("Expected %s to be of type %s but received %s.", name, expected, reflect.TypeOf(obj).String())
	}
	return nil
}

// parameterToString convert interface{} parameters to string, using a delimiter if format is provided.
func parameterToString(obj interface{}, collectionFormat string) string {
	var delimiter string

	switch collectionFormat {
	case "pipes":
		delimiter = "|"
	case "ssv":
		delimiter = " "
	case "tsv":
		delimiter = "\t"
	case "csv":
		delimiter = ","
	}

	if reflect.TypeOf(obj).Kind() == reflect.Slice {
		return strings.Trim(strings.Replace(fmt.Sprint(obj), " ", delimiter, -1), "[]")
	}

	return fmt.Sprintf("%v", obj)
}

// callAPI do the request.
func (c *APIClient) callAPI(request *http.Request) (*http.Response, error) {
	req, err := retryablehttp.FromRequest(request)
	if err != nil {
		return nil, err
	}
	return c.cfg.HTTPClient.Do(req)
}

// Change base path to allow switching to mocks
func (c *APIClient) ChangeBasePath(path string) {
	c.cfg.BasePath = path
}

// prepareRequest build the request
func (c *APIClient) prepareRequest(
	ctx context.Context,
	path string, method string,
	postBody interface{},
	headerParams map[string]string,
	queryParams url.Values,
	formParams url.Values,
	fileName string,
	fileBytes []byte) (localVarRequest *http.Request, err error) {

	var body *bytes.Buffer

	// Detect postBody type and post.
	if postBody != nil {
		contentType := headerParams["Content-Type"]
		if contentType == "" {
			contentType = detectContentType(postBody)
			headerParams["Content-Type"] = contentType
		}

		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}

	// add form parameters and file if available.
	if strings.HasPrefix(headerParams["Content-Type"], "multipart/form-data") && len(formParams) > 0 || (len(fileBytes) > 0 && fileName != "") {
		if body != nil {
			return nil, errors.New("Cannot specify postBody and multipart form at the same time.")
		}
		body = &bytes.Buffer{}
		w := multipart.NewWriter(body)

		for k, v := range formParams {
			for _, iv := range v {
				if strings.HasPrefix(k, "@") { // file
					err = addFile(w, k[1:], iv)
					if err != nil {
						return nil, err
					}
				} else { // form value
					w.WriteField(k, iv)
				}
			}
		}
		if len(fileBytes) > 0 && fileName != "" {
			w.Boundary()
			//_, fileNm := filepath.Split(fileName)
			part, err := w.CreateFormFile("file", filepath.Base(fileName))
			if err != nil {
				return nil, err
			}
			_, err = part.Write(fileBytes)
			if err != nil {
				return nil, err
			}
			// Set the Boundary in the Content-Type
			headerParams["Content-Type"] = w.FormDataContentType()
		}

		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
		w.Close()
	}

	if strings.HasPrefix(headerParams["Content-Type"], "application/x-www-form-urlencoded") && len(formParams) > 0 {
		if body != nil {
			return nil, errors.New("Cannot specify postBody and x-www-form-urlencoded form at the same time.")
		}
		body = &bytes.Buffer{}
		body.WriteString(formParams.Encode())
		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
	}

	// Setup path and query parameters
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Adding Query Param
	query := url.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	url.RawQuery = query.Encode()

	// Generate a new request
	if body != nil {
		localVarRequest, err = http.NewRequest(method, url.String(), body)
	} else {
		localVarRequest, err = http.NewRequest(method, url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers.Set(h, v)
		}
		localVarRequest.Header = headers
	}

	// Override request host, if applicable
	if c.cfg.Host != "" {
		localVarRequest.Host = c.cfg.Host
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.cfg.UserAgent)

	if ctx != nil {
		// add context to the request
		localVarRequest = localVarRequest.WithContext(ctx)

		// Walk through any authentication.

		// OAuth2 authentication
		if tok, ok := ctx.Value(ContextOAuth2).(oauth2.TokenSource); ok {
			// We were able to grab an oauth2 token from the context
			var latestToken *oauth2.Token
			if latestToken, err = tok.Token(); err != nil {
				return nil, err
			}

			latestToken.SetAuthHeader(localVarRequest)
		}

		// Basic HTTP Authentication
		if auth, ok := ctx.Value(ContextBasicAuth).(BasicAuth); ok {
			localVarRequest.SetBasicAuth(auth.UserName, auth.Password)
		}

		// AccessToken Authentication
		if auth, ok := ctx.Value(ContextAccessToken).(string); ok {
			localVarRequest.Header.Add("Authorization", "Bearer "+auth)
		}
	}

	for header, value := range c.cfg.DefaultHeader {
		localVarRequest.Header.Add(header, value)
	}

	return localVarRequest, nil
}

func (c *APIClient) decode(v interface{}, b []byte, contentType string) (err error) {
	if strings.Contains(contentType, "application/xml") {
		if err = xml.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	} else if strings.Contains(contentType, "application/json") {
		if err = json.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unsupported content type: %s", contentType)
}

// Add a file to the multipart request
func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}

// Prevent trying to import "fmt"
func reportError(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// Set request body from an interface{}
func setBody(body interface{}, contentType string) (bodyBuf *bytes.Buffer, err error) {
	if bodyBuf == nil {
		bodyBuf = &bytes.Buffer{}
	}

	if reader, ok := body.(io.Reader); ok {
		_, err = bodyBuf.ReadFrom(reader)
	} else if b, ok := body.([]byte); ok {
		_, err = bodyBuf.Write(b)
	} else if s, ok := body.(string); ok {
		_, err = bodyBuf.WriteString(s)
	} else if s, ok := body.(*string); ok {
		_, err = bodyBuf.WriteString(*s)
	} else if jsonCheck.MatchString(contentType) {
		err = json.NewEncoder(bodyBuf).Encode(body)
	} else if xmlCheck.MatchString(contentType) {
		xml.NewEncoder(bodyBuf).Encode(body)
	}

	if err != nil {
		return nil, err
	}

	if bodyBuf.Len() == 0 {
		err = fmt.Errorf("Invalid body type %s\n", contentType)
		return nil, err
	}
	return bodyBuf, nil
}

// detectContentType method is used to figure out `Request.Body` content type for request header
func detectContentType(body interface{}) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Ptr:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

// Ripped from https://github.com/gregjones/httpcache/blob/master/httpcache.go
type cacheControl map[string]string

func parseCacheControl(headers http.Header) cacheControl {
	cc := cacheControl{}
	ccHeader := headers.Get("Cache-Control")
	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyval := strings.Split(part, "=")
			cc[strings.Trim(keyval[0], " ")] = strings.Trim(keyval[1], ",")
		} else {
			cc[part] = ""
		}
	}
	return cc
}

// CacheExpires helper function to determine remaining time before repeating a request.
func CacheExpires(r *http.Response) time.Time {
	// Figure out when the cache expires.
	var expires time.Time
	now, err := time.Parse(time.RFC1123, r.Header.Get("date"))
	if err != nil {
		return time.Now()
	}
	respCacheControl := parseCacheControl(r.Header)

	if maxAge, ok := respCacheControl["max-age"]; ok {
		lifetime, err := time.ParseDuration(maxAge + "s")
		if err != nil {
			expires = now
		}
		expires = now.Add(lifetime)
	} else {
		expiresHeader := r.Header.Get("Expires")
		if expiresHeader != "" {
			expires, err = time.Parse(time.RFC1123, expiresHeader)
			if err != nil {
				expires = now
			}
		}
	}
	return expires
}

func strlen(s string) int {
	return utf8.RuneCountInString(s)
}

// GenericSwaggerError Provides access to the body, error and model on returned errors.
type GenericSwaggerError struct {
	body  []byte
	error string
	model interface{}
}

// Error returns non-empty string if there was an error.
func (e GenericSwaggerError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response
func (e GenericSwaggerError) Body() []byte {
	return e.body
}

// Model returns the unpacked model of the error
func (e GenericSwaggerError) Model() interface{} {
	return e.model
}
