package apps

import (
	"encoding/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

func portsToMap(ports interface{}) map[string]int {
	unmarshaled := map[string]interface{}{}
	must.OK(json.Unmarshal(must.Bytes(json.Marshal(ports)), &unmarshaled))

	res := map[string]int{}
	for k, v := range unmarshaled {
		res[k] = int(v.(float64))
	}
	return res
}
