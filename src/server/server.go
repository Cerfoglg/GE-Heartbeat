package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "bytes"
  "os"
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
    	http.Error(w, err.Error(), http.StatusBadRequest)
    	return
    }
    r.Body.Close()
    
    // Creating http client
    client := &http.Client{}
    
    // Creating keystone request for token
    data := []byte(`{"auth":{"passwordCredentials":{"username": "`+conf.USERNAME+`", "password":"`+conf.PASSWORD+`"}, "tenantName":"`+conf.TENANT+`"}}`)
    req, err := http.NewRequest("POST", ""+conf.KEYSTONE_HOST+":"+conf.KEYSTONE_PORT+"/v2.0/tokens", bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    
    // Requesting token
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Failed to contact Keystone: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if resp.StatusCode != 200 {
    	fmt.Println("Error with Keystone: ", resp.Status, resp.Header, resp.Body)
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    	return
    }
    
    // Extracting token
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Failed to read response from Keystone: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    var f interface{}
	err = json.Unmarshal(body, &f)
	if err != nil {
        fmt.Println("Failed to unmarshal JSON response from keystone: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	token := f.(map[string]interface{})["access"].(map[string]interface{})["token"].(map[string]interface{})["id"]
	if token == nil {
        fmt.Println("Failed to obtain token from keystone: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    resp.Body.Close()
    
    // Creating request for Monasca
    data = []byte(`{"name": "GE_Heartbeat", "dimensions": {"id": "`+m.ID+`", "enabler_id": "`+m.Enabler_ID+`", "enabler_version": "`+m.Enabler_Version+`"}, "timestamp": `+m.Timestamp+`, "value": 1}`)
    req, err = http.NewRequest("POST", ""+conf.MONASCA_HOST+":"+conf.MONASCA_PORT+"/v2.0/metrics", bytes.NewBuffer(data))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Auth-Token", token.(string))
    
    // Sending metric to Monasca
    resp, err = client.Do(req)
    if err != nil {
        fmt.Println("Failed to contact Monasca: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if resp.StatusCode == 204 {
	    w.WriteHeader(200)
	    return
    } else {
    	fmt.Println("Error with Monasca: ", resp.Status, resp.Header, resp.Body)
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    	return
    }
    resp.Body.Close()
}

func setConfEnvVariables() {
	if os.Getenv("KEYSTONE_HOST") != "" {
		conf.KEYSTONE_HOST = os.Getenv("KEYSTONE_HOST")
	}
	if os.Getenv("KEYSTONE_PORT") != "" {
		conf.KEYSTONE_HOST = os.Getenv("KEYSTONE_PORT")
	}
	if os.Getenv("MONASCA_HOST") != "" {
		conf.KEYSTONE_HOST = os.Getenv("MONASCA_HOST")
	}
	if os.Getenv("MONASCA_PORT") != "" {
		conf.KEYSTONE_HOST = os.Getenv("MONASCA_PORT")
	}
	if os.Getenv("USERNAME") != "" {
		conf.KEYSTONE_HOST = os.Getenv("USERNAME")
	}
	if os.Getenv("PASSWORD") != "" {
		conf.KEYSTONE_HOST = os.Getenv("PASSWORD")
	}
	if os.Getenv("TENANT") != "" {
		conf.KEYSTONE_HOST = os.Getenv("TENANT")
	}
}
 
func main() {
	// Reading configuration
	conf = config{}
	data, err := ioutil.ReadFile("configuration.yml")
	if err != nil {
        fmt.Println("Failed to read configuration: ", err)
        os.Exit(1)
    }
    err = yaml.Unmarshal([]byte(data), &conf)
    if err != nil {
	    fmt.Println("Failed to unmarshal configuration: ", err)
	    os.Exit(1)
    }
    setConfEnvVariables()
	
    http.HandleFunc("/beat", handler)
    http.ListenAndServe(":8080", nil)
}