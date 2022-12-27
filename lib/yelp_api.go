package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Search(location string, term string, categories string) map[string]interface{} {
	apiKey := fmt.Sprintf("Bearer %s", os.Getenv("YELP_API_KEY"))
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.yelp.com/v3/businesses/search", nil)
	q := req.URL.Query()
	q.Add("term", term)
	q.Add("location", location)
	q.Add("categories", categories)
	parameters := map[string]string{
		"sort_by":  "best_match",
		"open_now": "true",
		"radius":   "16000",
		"limit":    "5",
	}
	for k, v := range parameters {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", apiKey)
	res, _ := client.Do(req)

	defer res.Body.Close()

	type ApiRespond struct {
		businesses []interface{}
	}

	b, _ := io.ReadAll(res.Body)
	var jsonRes map[string]interface{}
	_ = json.Unmarshal(b, &jsonRes)
	businesses := jsonRes["businesses"].([]interface{})

	returnMap := make(map[string]interface{})

	for _, element := range businesses {
		// fmt.Println(element)
		var id = element.(map[string]interface{})["id"].(string)
		keys := []string{"name", "location", "image_url", "url", "display_phone", "rating", "review_count"}
		selectedMap := make(map[string]interface{})
		for _, key := range keys {
			selectedMap[key] = element.(map[string]interface{})[key]
		}
		returnMap[id] = selectedMap
	}
	return returnMap
}
