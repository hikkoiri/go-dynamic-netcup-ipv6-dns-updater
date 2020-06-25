package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	customerNr  string
	apiKey      string
	apiPassword string
	domain      string
	hosts       []string
	apiUrl      string
)

func main() {
	fmt.Println("Starting Netcup IPv6 updater at:", time.Now())

	fmt.Println("Checking environment variables")
	checkEnvVariables()

	fmt.Println("Get IPv6 Address")
	ipv6 := getIPv6Address()
	fmt.Println("Found public IPv6 Address", ipv6)

	fmt.Println("Netcup Login")
	apiSessionId := login()
	fmt.Println("Successfully logged in.")

	fmt.Println("Update DNS records.")
	updateDNSRecords(hosts, apiSessionId, ipv6)

	fmt.Println("Netcup Logout")
	logout(apiSessionId)

	fmt.Println("Updater finished successfully.")

}

func updateDNSRecords(hosts []string, apiSessionId string, currentIpV6 string) {

	//retrieve dns records
	dnsRecords := getDNSRecords(apiSessionId)

	for _, host := range hosts {
		fmt.Println("Updating host:", host)

		//check if record exists
		doesRecordExist := false
		var foundDnsRecord dnsRecord
		for _, extractedDnsRecord := range dnsRecords {
			if host == extractedDnsRecord.Hostname {
				if extractedDnsRecord.Type == "AAAA" {
					doesRecordExist = true
					foundDnsRecord = extractedDnsRecord
					break
				}
			}
		}



		if doesRecordExist {
			// check if retrieved ipv6 is similar to current one
			if foundDnsRecord.Destination == currentIpV6 {
				fmt.Println("DNS record already exist. destination did not change.")
				//nice, nothing to do
			} else {
				fmt.Println("DNS record already exist. Updating", foundDnsRecord.Destination, "to", currentIpV6)

				foundDnsRecord.Destination = currentIpV6

				updateExistingDnsRecord(apiSessionId, foundDnsRecord)
			}
		} else {
			//create new ipv6 record

			dnsRecordCandidate := map[string]interface{}{
				"hostname":    host,
				"type":        "AAAA",
				"destination": currentIpV6,
			}

			fmt.Println("DNS record does not exist. Creating one.")
			createNewDnsRecords(apiSessionId, dnsRecordCandidate)
		}
	}
}

func updateExistingDnsRecord(apiSessionId string, updatedDnsRecord dnsRecord) {
	//delete old record

	oldRecordToDelete := map[string]interface{}{
		"id":           updatedDnsRecord.Id,
		"hostname":     updatedDnsRecord.Hostname,
		"type":         updatedDnsRecord.Type,
		"destination":  updatedDnsRecord.Destination,
		"deleterecord": true,
	}
	createNewDnsRecords(apiSessionId,oldRecordToDelete)

	//create new one
	dnsRecordToUpdate := map[string]interface{}{
		"hostname":     updatedDnsRecord.Hostname,
		"type":         updatedDnsRecord.Type,
		"destination":  updatedDnsRecord.Destination,
	}
	createNewDnsRecords(apiSessionId,dnsRecordToUpdate)
}

func createNewDnsRecords(apiSessionId string, newDnsRecord map[string]interface{}) {
	requestBody := map[string]interface{}{
		"action": "updateDnsRecords",
		"param": map[string]interface{}{
			"domainname":     domain,
			"customernumber": customerNr,
			"apikey":         apiKey,
			"apisessionid":   apiSessionId,
			"dnsrecordset": map[string][]map[string]interface{}{
				"dnsrecords": {newDnsRecord},
			},
		},
	}
	response, success := netcupApiPost(requestBody)
	if !success {
		fmt.Println("Exiting. Could not create or update DNS records.")
		os.Exit(1)
	}

	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(response)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	message, _ := jq.String("longmessage")
	fmt.Println("Update response:", message)

}

type dnsRecord struct {
	Id           string
	Hostname     string
	Type         string
	Priority     string
	Destination  string
	Deleterecord bool
	State        string
}

func getDNSRecords(apiSessionId string) []dnsRecord {
	requestBody := map[string]interface{}{
		"action": "infoDnsRecords",
		"param": map[string]string{
			"domainname":     domain,
			"customernumber": customerNr,
			"apikey":         apiKey,
			"apisessionid":   apiSessionId,
		},
	}

	response, success := netcupApiPost(requestBody)
	if !success {
		fmt.Println("Exiting. Could not get DNS records.")
		os.Exit(1)
	}

	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(response)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	dnsRecords, _ := jq.ArrayOfObjects("responsedata", "dnsrecords")

	//dnsRecords2 := dnsRecords.([]dnsRecord)
	//does not work as expected -> going the long way

	var resolvedDnsRecords []dnsRecord
	for _, extractedDnsRecord := range dnsRecords {
		dnsRecordToAdd := dnsRecord{
			Id:           extractedDnsRecord["id"].(string),
			Hostname:     extractedDnsRecord["hostname"].(string),
			Type:         extractedDnsRecord["type"].(string),
			Priority:     extractedDnsRecord["priority"].(string),
			Destination:  extractedDnsRecord["destination"].(string),
			Deleterecord: extractedDnsRecord["deleterecord"].(bool),
			State:        extractedDnsRecord["state"].(string),
		}
		resolvedDnsRecords = append(resolvedDnsRecords, dnsRecordToAdd)
	}
	return resolvedDnsRecords
}

func logout(apiSessionId string) {

	requestBody := map[string]interface{}{
		"action": "logout",
		"param": map[string]string{
			"customernumber": customerNr,
			"apikey":         apiKey,
			"apisessionid":   apiSessionId,
		},
	}

	response, success := netcupApiPost(requestBody)
	if !success {
		fmt.Println("Exiting. Could not logout.")
		os.Exit(1)
	}

	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(response)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)

	errorMessage, err := jq.String("longmessage")
	if err != nil {
		fmt.Println("Could not even read logout response message from logout response.")
	} else {
		fmt.Println("API Logout response:", errorMessage)
	}
}

func login() string {

	requestBody := map[string]interface{}{
		"action": "login",
		"param": map[string]string{
			"customernumber": customerNr,
			"apikey":         apiKey,
			"apipassword":    apiPassword,
		},
	}

	response, success := netcupApiPost(requestBody)
	if !success {
		fmt.Println("Exiting. Could not login.")
		os.Exit(1)
	}

	//try to find apisessionkey
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(response)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)

	apisessionid, err := jq.String("responsedata", "apisessionid")
	if err != nil {
		fmt.Println("Could not retrieve ApiSessionKey for further API requests.")
		errorMessage, err := jq.String("longmessage")
		if err != nil {
			fmt.Println("Could not even read error message from login response.")
		} else {
			fmt.Println("Error message:", errorMessage)
		}
		os.Exit(1)
	}

	return apisessionid
}

func getIPv6Address() string {
	url1 := "https://ip6.seeip.org"
	url2 := "https://v6.ident.me/"

	fmt.Println("Fetching public IPv6 address from", url1)
	ipv6, success := httpGet(url1)
	if !success {
		fmt.Println("Retrying with url", url2)
		ipv6, success := httpGet(url2)
		if !success {
			fmt.Println("Could not determine public IPv6 address. Exiting")
			os.Exit(1)
		}
		return ipv6
	}
	return ipv6
}

func checkEnvVariables() {
	customerNr = os.Getenv("CUSTOMERNR")
	if len(customerNr) != 0 {
		fmt.Println("CUSTOMERNR set")

	} else {
		fmt.Println("CUSTOMERNR not set. Exiting")
		os.Exit(1)
	}

	apiKey = os.Getenv("APIKEY")
	if len(apiKey) != 0 {
		fmt.Println("APIKEY set")

	} else {
		fmt.Println("APIKEY not set. Exiting")
		os.Exit(1)
	}

	apiPassword = os.Getenv("APIPASSWORD")
	if len(apiPassword) != 0 {
		fmt.Println("APIPASSWORD set")

	} else {
		fmt.Println("APIPASSWORD not set. Exiting")
		os.Exit(1)
	}

	domain = os.Getenv("DOMAIN")
	if len(domain) != 0 {
		fmt.Println("DOMAIN set")
	} else {
		fmt.Println("DOMAIN not set. Exiting")
		os.Exit(1)
	}

	hostsAsString := os.Getenv("HOSTS")
	if len(hostsAsString) != 0 {
		fmt.Println("HOSTS set")
		fmt.Println("Splitting comma separated string HOSTS")
		hosts = strings.Split(hostsAsString, ",")
		if len(hosts) > 0 {
			fmt.Println("success")
		} else {
			fmt.Println("failure. lists of hosts is empty")
			os.Exit(1)
		}
	} else {
		fmt.Println("HOSTS not set. Exiting")
		os.Exit(1)
	}

	apiUrl = getEnvWithFallBack("APIURL", "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON")
}

func getEnvWithFallBack(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		fmt.Println(key, "not set. Falling back to", fallback)
		return fallback
	}
	return value
}

func httpGet(url string) (string, bool) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("An error occurred executing the http request", err.Error())
		return "", false
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("An error occurred reading the response body", err.Error())
			return "", false
		}
		return string(body), true
	}
}

func netcupApiPost(requestBody interface{}) ([]byte, bool) {

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Exiting. Unexpected JSON marshaling error occurred:", err)
		os.Exit(1)
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("An error occurred executing the http request", err.Error())
		return nil, false
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("An error occurred reading the response body", err.Error())
			return nil, false
		}
		return body, true
	}
}
