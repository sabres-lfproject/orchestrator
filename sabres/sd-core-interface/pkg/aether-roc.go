package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/onosproject/aether-roc-api/pkg/aether_2_1_0/types"
	// log "github.com/sirupsen/logrus"
)

func GetEnterprises(endpoint string) ([]types.EnterpriseId, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/aether-roc-api/targets", endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("resp: %s", string(body))

	var respJson []map[string]string
	err = json.Unmarshal(body, &respJson)

	ei := make([]types.EnterpriseId, 0)
	for _, ent := range respJson {
		name, ok := ent["name"]
		if ok {
			ei = append(ei, types.EnterpriseId(name))
		}
	}

	return ei, err
}

func GetSites(endpoint, enterprise string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site", endpoint, enterprise))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("resp: %s", string(body))

	var respJson []map[string]interface{}
	err = json.Unmarshal(body, &respJson)

	si := make([]string, 0)
	for _, ent := range respJson {
		sid, ok2 := ent["site-id"].(string)
		if ok2 {
			si = append(si, sid)
		}
	}

	return si, err
}

func GetSiteDetails(endpoint, enterprise, site, detail string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s", endpoint, enterprise, site))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJson map[string]interface{}
	err = json.Unmarshal(body, &respJson)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("resp: %#v\n", respJson)
	fmt.Printf("resp: %#v\n", respJson[detail])

	_, ok := respJson[detail]
	if !ok {
		return nil, fmt.Errorf("Key not found: %s", detail)
	}

	var si map[string]interface{}
	si = make(map[string]interface{})

	switch detail {
	case "device-group":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["device-group-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device group without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	case "device":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["device-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	case "ip-domain":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["ip-domain-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	case "sim-card":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["sim-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	case "slice":
		// SiteSlice
		//tmp := respJson[detail].(map[string][]interface{}).(types.SiteSliceList)
		tmp := respJson[detail].([]interface{})
		//si = respJson[detail].(map[string]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			//ssl := ttmpx.(types.SiteSlice)
			//sid := string(ssl.SliceId)
			sid := ttmpx["slice-id"].(string)
			if sid == "" {
				return nil, fmt.Errorf("slice without slice id")
			}
			//si[sid] = ssl
			si[sid] = ttmpx
		}
		break
	case "small-cell":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["small-cell-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	case "upf":
		tmp := respJson[detail].([]interface{})
		for _, tmpx := range tmp {
			ttmpx := tmpx.(map[string]interface{})
			dgid := ttmpx["upf-id"].(string)
			if dgid == "" {
				return nil, fmt.Errorf("device without identifier")
			}
			si[dgid] = ttmpx
		}
		break
	default:
		return nil, fmt.Errorf("Unknown key parameter: %s", detail)
	}

	return si, nil
}

func CreateSite(endpoint, enterprise, site string) (string, error) {
	desc := "Created from cli"
	dn := fmt.Sprintf("%s-cli", site)
	data := &types.Site{
		Description: &desc,
		SiteId:      types.ListKey(site),
		DisplayName: &dn,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	ep := fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s", endpoint, enterprise, site)
	request, err := http.NewRequest("POST", ep, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("resp: %s", string(body))

	return string(body), err
}

func CreateSiteFromFile(endpoint, enterprise, site, fiName string) (string, error) {
	return CreatePatchFromFile(endpoint, enterprise, site, "site", fiName)
}

func CreateDeviceGroupFromFile(endpoint, enterprise, site, fiName string) (string, error) {
	return CreatePatchFromFile(endpoint, enterprise, site, "device-group", fiName)
}

func CreateDevicesFromFile(endpoint, enterprise, site, fiName string) (string, error) {
	return CreatePatchFromFile(endpoint, enterprise, site, "device", fiName)
}

func CreateSliceFromFile(endpoint, enterprise, site, fiName string) (string, error) {
	return CreatePatchFromFile(endpoint, enterprise, site, "slice", fiName)
}

type SiteObj struct {
	AdditionalProperties map[string]string          `json:"additionalProperties,omitempty"`
	SiteName             string                     `json:"site-id,omitempty"`
	Site                 *types.SiteList            `json:"site,omitempty"`
	DeviceGroup          *types.SiteDeviceGroupList `json:"device-group,omitempty"`
	Slice                *types.SiteSliceList       `json:"slice,omitempty"`
	Device               *types.SiteDeviceList      `json:"device,omitempty"`
}

type UpdateObj struct {
	Protocol []*SiteObj `json:"site-2.1.0,omitempty"`
}

type DeleteObj struct{}

type PatchObj struct {
	Target  string     `json:"default-target,omitempty"`
	Updates *UpdateObj `json:"Updates,omitempty"`
	Deletes *DeleteObj `json:"Deletes,omitempty"`
}

func CreatePatchFromFile(endpoint, ent, site, fit, fin string) (string, error) {
	siteDataList, err := ReadXFromFile(fit, fin)
	if err != nil {
		return "", err
	}

	po := &PatchObj{
		Target: "default-ent",
		Updates: &UpdateObj{
			Protocol: []*SiteObj{
				&SiteObj{
					AdditionalProperties: map[string]string{
						"enterprise-id": ent,
					},
					SiteName: site,
				},
			},
		},
	}

	do := po.Updates.Protocol[0]
	switch fit {
	case "site":
		sL := make(types.SiteList, 0)
		for _, site := range siteDataList {
			sL = append(sL, *site)
		}
	case "device-group":
		do.DeviceGroup = siteDataList[0].DeviceGroup
	case "device":
		do.Device = siteDataList[0].Device
	case "slice":
		do.Slice = siteDataList[0].Slice
	default:
		return "", fmt.Errorf("should never be executed")
	}

	ep := fmt.Sprintf("http://%s/aether-roc-api/aether-roc-api", endpoint)

	patchData, err := json.Marshal(po)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("PATCH", ep, bytes.NewBuffer(patchData))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("resp: %s", string(result))

	return string(result), err
}

func ReadXFromFile(fiType, fiName string) ([]*types.Site, error) {

	body, err := ioutil.ReadFile(fiName)
	if err != nil {
		return nil, err
	}

	returnData := make([]*types.Site, 0)
	returnData = append(returnData, &types.Site{})

	switch fiType {
	case "site":
		// https://github.com/onosproject/aether-roc-api/blob/master/pkg/aether_2_1_0/types/aether-2.1.0-types.go#L124
		var data *types.Site
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		returnData[0] = data
		return returnData, nil
	case "device-group":
		// https://github.com/onosproject/aether-roc-api/blob/master/pkg/aether_2_1_0/types/aether-2.1.0-types.go#L260
		var data *types.SiteDeviceGroup
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		returnData[0].DeviceGroup = &types.SiteDeviceGroupList{*data}
		return returnData, nil
	case "device":
		// https://github.com/onosproject/aether-roc-api/blob/master/pkg/aether_2_1_0/types/aether-2.1.0-types.go#L274
		var data *types.SiteDeviceList
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		returnData[0].Device = data
		return returnData, nil
	case "slice":
		// https://github.com/onosproject/aether-roc-api/blob/master/pkg/aether_2_1_0/types/aether-2.1.0-types.go#L481
		var data *types.SiteSlice
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		fmt.Printf("slice: %v\n", data)
		returnData[0].Slice = &types.SiteSliceList{*data}
		return returnData, nil
	default:
		return nil, fmt.Errorf("Unknown file type to read")
	}

	return returnData, nil
}

/*
func CreateXFromFile(endpoint, enterprise, site, fiType, objName, fiName string) (string, error) {

	jsonData, err := ReadXFromFile(fiType, fiName)
	if err != nil {
		return "", err
	}

	ep := ""
	if fiType == "site" {
		ep = fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s", endpoint, enterprise, site)
	} else {
		ep = fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s/%s/%s", endpoint, enterprise, site, fiType, objName)
	}

	request, err := http.NewRequest("POST", ep, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("resp: %s", string(result))

	return string(result), err
}
*/

func DeleteAetherObject(endpoint, enterprise, site, objType, objName string) (string, error) {

	ep := ""
	if objType == "site" {
		ep = fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s", endpoint, enterprise, site)
	} else {
		ep = fmt.Sprintf("http://%s/aether-roc-api/aether/v2.1.x/%s/site/%s/%s/%s", endpoint, enterprise, site, objType, objName)
	}

	request, err := http.NewRequest("DELETE", ep, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("resp: %s", string(result))

	return string(result), err
}
