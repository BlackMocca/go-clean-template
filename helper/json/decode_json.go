package helper

import "encoding/json"

func GetParamsFromJsonData(postForm map[string][]string) interface{} {
	var data interface{}
	var dataString = postForm["data"][0]
	json.Unmarshal([]byte(dataString), &data)
	return data
}
