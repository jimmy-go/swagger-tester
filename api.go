package openapitester

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// API type contains all swagger info.
type API struct {
	Host        string                                `json:"host"`
	Paths       map[string]map[string]json.RawMessage `json:"paths"`
	Definitions map[string]*Definition                `json:"definitions"`
	Schemes     []string                              `json:"schemes"`
}

// Domain returns the scheme and host. It defaults to https if present.
func (a *API) Domain() string {
	var scheme string
	for _, s := range a.Schemes {
		if s == "https" {
			return s + "://" + a.Host
		}
		scheme = s
	}
	return scheme + "://" + a.Host
}

// Search searchs method and request uri skipping url params as '/path/*/something'.
func (a *API) Search(method, requestURI string) (*PathMethod, error) {
	for k, uris := range a.Paths {
		kc := removeVars(k)
		rc := removeVars(requestURI)
		if kc == rc {
			for method2, v := range uris {
				if strings.ToUpper(method2) == strings.ToUpper(method) {
					if method2 == "parameters" {
						continue
					}
					var p *PathMethod
					if err := json.Unmarshal(v, &p); err != nil {
						return nil, err
					}
					return p, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("not found: %s %s", method, requestURI)
}

// Examples returns examples bodies if resource exists.
func (a *API) Examples(method, requestURI string) ([]string, error) {
	pm, err := a.Search(method, requestURI)
	if err != nil {
		return nil, err
	}
	var res []string
	for i := range pm.Parameters {
		x := pm.Parameters[i]
		if x.Schema == nil {
			continue
		}
		ref := strings.Replace(x.Schema.Ref, "#/definitions/", "", -1)
		def, ok := a.Definitions[ref]
		if !ok {
			continue
		}
		res = append(res, def.Example)
	}
	if len(res) == 0 {
		return nil, errors.New("example not found")
	}
	return res, nil
}
