package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "bytes"
  "io/ioutil"
  "gopkg.in/yaml.v2"
)

// Config struct
type config struct {
	KEYSTONE_HOST string
	KEYSTONE_PORT string
	MONASCA_HOST string
	MONASCA_PORT string
	USERNAME string
	PASSWORD string
	TENANT string
}

// Config var
var conf config

// Struct of the message sent by the Heartbeat scripts
type message struct {
    ID string
    Enabler_ID string
    Enabler_Version string
    Timestamp string
}

// Main handler function
func handler(w http.ResponseWriter, r *http.Request) {
	// Decoding the message sent by the scripts
    decoder := json.NewDecoder(r.Body)
    var m message   
    err := decoder.Decode(&m)
    if err != nil {
    	fmt.Println("Cannot decode message: ", err)
    	return
    }
    fmt.Println(m.ID)
    r.Body.Close()
    
    // Authenticating with Keystone and sending metric
    client := &http.Client{}
    
    data := []byte(`{"auth":{"passwordCredentials":{"username": "`+conf.USERNAME+`", "password":"`+conf.PASSWORD+`"}, "tenantName":"`+conf.TENANT+`"}}`)
    req, err := http.NewRequest("POST", ""+conf.KEYSTONE_HOST+":"+conf.KEYSTONE_PORT+"/v2.0/tokens", bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Failed to contact Keystone: ", err)
        return
    }
    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Failed to read response from Keystone: ", err)
        return
    }
    var f interface{}
	err = json.Unmarshal(body, &f)
	token := f.(map[string]interface{})["access"].(map[string]interface{})["token"].(map[string]interface{})["id"]
	fmt.Println("response Body:", token)
    resp.Body.Close()
    
    data = []byte(`{"name": "GE_Heartbeat", "dimensions": {"id": "`+m.ID+`", "enabler_id": "`+m.Enabler_ID+`", "enabler_version": "`+m.Enabler_Version+`"}, "timestamp": `+m.Timestamp+`, "value": 1}`)
    req, err = http.NewRequest("POST", ""+conf.MONASCA_HOST+":"+conf.MONASCA_PORT+"/v2.0/metrics", bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Auth-Token", token.(string))
    
    resp, err = client.Do(req)
    if err != nil {
        fmt.Println("Failed to contact Monasca: ", err)
        return
    }
    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
}
 
func main() {
	// Reading configuration
	conf = config{}
	data, err := ioutil.ReadFile("configuration.yml")
	if err != nil {
        fmt.Println("Failed to read configuration: ", err)
        return
    }
    err = yaml.Unmarshal([]byte(data), &conf)
    if err != nil {
	    fmt.Println("Failed to unmarshal configuration: ", err)
        return
    }
	
    http.HandleFunc("/beat", handler)
    http.ListenAndServe(":8080", nil)
}