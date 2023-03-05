package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	targetAPI = "https://api.wasi.co"
)

type Response struct {
	Items []map[string]interface{} `json:"items"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := url.Parse(targetAPI + r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusInternalServerError)
			return
		}
		targetURL.RawQuery = r.URL.RawQuery

		targetReq, err := http.NewRequest(r.Method, targetURL.String(), nil)
		if err != nil {
			log.Fatal(err)
		}
		targetResp, err := http.DefaultClient.Do(targetReq)
		if err != nil {
			http.Error(w, "Error forwarding request to target", http.StatusInternalServerError)
			return
		}
		defer targetResp.Body.Close()

		defer targetResp.Body.Close()
		var data map[string]json.RawMessage
		err = json.NewDecoder(targetResp.Body).Decode(&data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var items []map[string]interface{}
		for key, value := range data {
			if key == "total" {
				continue
			}
			if key == "status" {
				continue
			}
			var item map[string]interface{}
			err := json.Unmarshal(value, &item)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if strings.Contains(r.URL.Path, "/v1/property") {
				var images []interface{}
				for gallery := range item["galleries"].([]interface{}) {
					for iKey, image := range item["galleries"].([]interface{})[gallery].(map[string]interface{}) {
						if iKey == "id" {
							continue
						}
						images = append(images, image)
					}
				}
				delete(item, "galleries")
				item["images"] = images
			}

			items = append(items, item)
		}

		transformedResponse := Response{Items: items}
		transformedJSON, err := json.Marshal(transformedResponse)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Return as JSON
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, "%v", string(transformedJSON))
	})

	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
