package hms_client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	base "github.com/Cray-HPE/hms-base/v2"
)

type HSMClient struct {
	baseURL      string
	instanceName string
	client       *http.Client
	apiToken     string
}

func NewHSMClient(baseURL string, client *http.Client, instanceName string) *HSMClient {
	return &HSMClient{
		baseURL:      baseURL,
		client:       client,
		instanceName: instanceName,
	}
}

func (sc *HSMClient) WithAPIToken(apiToken string) *HSMClient {
	sc.apiToken = apiToken
	return sc
}

func (sc *HSMClient) addAPITokenHeader(request *http.Request) {
	if sc.apiToken != "" {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sc.apiToken))
	}
}

func (sc *HSMClient) GetServiceValues(ctx context.Context) (HMSValues, error) {
	// Build up the request
	request, err := http.NewRequestWithContext(ctx, "GET", sc.baseURL+"/v2/service/values", nil)
	if err != nil {
		return HMSValues{}, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	if err != nil {
		return HMSValues{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return HMSValues{}, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var serviceValues HMSValues
	if err := json.NewDecoder(response.Body).Decode(&serviceValues); err != nil {
		return HMSValues{}, err
	}

	return serviceValues, nil
}

type StateComponentsSearchFilter struct {
	NID     []int
	Class   []string
	Role    []string
	SubRole []string
	State   []string
	// Enabled []bool
}

func (filter *StateComponentsSearchFilter) String() string {
	params := url.Values{}

	for _, nid := range filter.NID {
		params.Set("nid", fmt.Sprintf("%d", nid))
	}
	for _, class := range filter.Class {
		params.Set("class", class)
	}
	for _, role := range filter.Role {
		params.Set("role", role)
	}
	for _, subRole := range filter.SubRole {
		params.Set("subrole", subRole)
	}
	for _, state := range filter.State {
		params.Set("state", state)
	}

	return params.Encode()
}

func (sc *HSMClient) GetStateComponents(ctx context.Context) (base.ComponentArray, error) {
	return sc.GetStateComponentsFilter(ctx, &StateComponentsSearchFilter{})
}

func (sc *HSMClient) GetStateComponentsFilter(ctx context.Context, filter *StateComponentsSearchFilter) (base.ComponentArray, error) {
	// Build up the request
	url := sc.baseURL + "/v2/State/Components"
	if filter != nil {
		url += "?" + filter.String()
	}

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return base.ComponentArray{}, err
	}
	base.SetHTTPUserAgent(request, sc.instanceName)
	sc.addAPITokenHeader(request)

	// Perform the request!
	response, err := sc.client.Do(request)
	if err != nil {
		return base.ComponentArray{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return base.ComponentArray{}, fmt.Errorf("unexpected status code %d expected 200", response.StatusCode)
	}

	var serviceValues base.ComponentArray
	if err := json.NewDecoder(response.Body).Decode(&serviceValues); err != nil {
		return base.ComponentArray{}, err
	}

	return serviceValues, nil
}
