package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type GetMachinesResp struct {
	Machines []struct {
		ID          string   `json:"id"`
		MachineKey  string   `json:"machineKey"`
		NodeKey     string   `json:"nodeKey"`
		DiscoKey    string   `json:"discoKey"`
		IPAddresses []string `json:"ipAddresses"`
		Name        string   `json:"name"`
		Namespace   struct {
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"namespace"`
		LastSeen             time.Time `json:"lastSeen"`
		LastSuccessfulUpdate time.Time `json:"lastSuccessfulUpdate"`
		Expiry               time.Time `json:"expiry"`
		PreAuthKey           struct {
			Namespace  string    `json:"namespace"`
			ID         string    `json:"id"`
			Key        string    `json:"key"`
			Reusable   bool      `json:"reusable"`
			Ephemeral  bool      `json:"ephemeral"`
			Used       bool      `json:"used"`
			Expiration time.Time `json:"expiration"`
			CreatedAt  time.Time `json:"createdAt"`
		} `json:"preAuthKey"`
		CreatedAt      time.Time `json:"createdAt"`
		RegisterMethod string    `json:"registerMethod"`
	} `json:"machines"`
}

type GetMachinesOptions struct {
	ServerURL string
	Namespace string
	APIKey    string
}

func GetMachines(client *http.Client, opts *GetMachinesOptions) (*GetMachinesResp, error) {
	u, err := url.Parse(opts.ServerURL + "/api/v1/machine")
	if err != nil {
		return nil, err
	}
	u.RawQuery = url.Values{
		"namespace": {opts.Namespace},
	}.Encode()
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"Authorization": {"Bearer " + opts.APIKey},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to GET %v: %w", u, err)
	}
	if resp.StatusCode != http.StatusOK {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("failed to GET %v with status %v: %s", u, resp.StatusCode, body.String())
	}
	respBody := &GetMachinesResp{}
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body as JSON: %w", err)
	}
	return respBody, nil
}

func main() {
	checkIntervalStr := os.Getenv(("HEADSCALE2HOSTS_CHECK_INTERVAL"))
	if checkIntervalStr == "" {
		checkIntervalStr = "1m"
	}
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		log.Fatal("failed to parse HEADSCALE2HOSTS_CHECK_INTERVAL: ", err)
	}
	serverURL := os.Getenv("HEADSCALE2HOSTS_SERVER_URL")
	if serverURL == "" {
		log.Fatal("HEADSCALE2HOSTS_SERVER_URL is not set")
	}
	apiKey := os.Getenv("HEADSCALE2HOSTS_API_KEY")
	if apiKey == "" {
		log.Fatal("HEADSCALE2HOSTS_API_KEY is not set")
	}
	namespace := os.Getenv("HEADSCALE2HOSTS_NAMESPACE")
	if namespace == "" {
		log.Fatal("HEADSCALE2HOSTS_NAMESPACE is not set")
	}
	hostsFilePath := os.Getenv("HEADSCALE2HOSTS_HOSTS_FILE")
	if hostsFilePath == "" {
		hostsFilePath = "hosts"
	}
	domainSuffix := os.Getenv("HEADSCALE2HOSTS_DOMAIN_SUFFIX")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	opts := &GetMachinesOptions{
		ServerURL: serverURL,
		Namespace: namespace,
		APIKey:    apiKey,
	}
	for {
		resp, err := GetMachines(client, opts)
		if err != nil {
			log.Println("failed to get machines: ", err)
			time.Sleep(checkInterval)
			continue
		}
		hosts := &bytes.Buffer{}
		hosts.WriteString("# Generated by HEADSCALE2HOSTS\n")
		hosts.WriteString("# Generated at " + time.Now().Format(time.RFC3339) + "\n")
		hosts.WriteString("# Do not edit this file manually\n")
		hosts.WriteString("\n")
		maxIpLength := 0
		for _, machine := range resp.Machines {
			for _, ip := range machine.IPAddresses {
				if len(ip) > maxIpLength {
					maxIpLength = len(ip)
				}
			}
		}

		for _, machine := range resp.Machines {
			for _, ip := range machine.IPAddresses {
				hosts.WriteString(ip + strings.Repeat(" ", maxIpLength-len(ip)) + " " + machine.Name + domainSuffix + "\n")
			}
		}
		err = ioutil.WriteFile(hostsFilePath, hosts.Bytes(), 0644)
		if err != nil {
			log.Println("failed to write hosts file: ", err)
		}
		log.Printf("wrote %d entries to %s", len(resp.Machines), hostsFilePath)
		time.Sleep(checkInterval)
	}

}
