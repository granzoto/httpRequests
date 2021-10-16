package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	PRIVCONSOLE = "https://10.96.37.156:8080"
	PUBCONSOLE  = "https://10.98.79.157:8080"
	ADMUSER = "admin"
	ADMPASS = "zChhR6NPgC"
	)

type TokenState struct {
	Name            string     `json:"name"`
	ClaimsMade      *int       `json:"claimsMade"`
	ClaimsRemaining *int       `json:"claimsRemaining"`
	ClaimExpiration *time.Time `json:"claimExpiration"`
	Created         string     `json:"created,omitempty"`
}

func accessConsole(method string, url string, path string, body io.Reader, user string, pass string) (string, error) {

	// Define the request first
	req, err := http.NewRequest(method, url+"/"+path, body)
	if err != nil {
		return "", fmt.Errorf("Error defining the request")
	}

	// Define request basic auth
	if user != "" {
		req.SetBasicAuth(user, pass)
	}

	// Define the HTTP Client
	client := http.Client{}

	if strings.HasPrefix(url, "https") {
		// Accept insecure connections
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request")
	}

	bodyResp, err := ioutil.ReadAll(resp.Body)

	fmt.Println("Resp Body => ", string(bodyResp))

	return string(bodyResp), nil
}

func runCmd(cmd string, args ...string) bool {

	cmdToRun := exec.Command(cmd, args...)
	err := cmdToRun.Run()

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func main() {

	fmt.Println("Here we go !!")

	fmt.Println("Private Console")
	dataPriv, err := accessConsole("GET", PRIVCONSOLE, "DATA", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("data from private => ", dataPriv)
	}

	fmt.Println("Public Console")
	dataPub, err := accessConsole("GET", PUBCONSOLE, "DATA", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("data from private => ", dataPub)
	}

	// Create a claim token manually in west - public"
	ok := runCmd("skupper", "token", "create", "/tmp/token-west-renato01.yaml", "--expiry", "25m0s", "--name", "renato01", "--password", "rena-senha", "--token-type", "claim", "--uses",  "2", "-n", "west")
	if !ok {
		fmt.Println("Error to create the claim for public")
	}

	// Tokens from pub
	tokenPub, err := accessConsole("GET", PUBCONSOLE, "tokens", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("data from private => ", tokenPub)
	}



}
