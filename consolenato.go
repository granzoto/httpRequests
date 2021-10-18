package main

import (
	//"bytes"
	"crypto/tls"
	"encoding/json"
	"net/url"

	//"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

type TokenOptions struct {
	Expiry time.Duration	`json:"expiration"`
	Uses   int				`json:"uses"`
}

func accessConsole(method string, url string, path string, body io.Reader, user string, pass string) (string, error) {

	// Define the request first
	req, err := http.NewRequest(method, url+"/"+path, body)
	if err != nil {
		return "", fmt.Errorf("Error defining the request")
	}

	// If this is a POST
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
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

	strResp := string(bodyResp)

	//fmt.Println("Resp Body => ", strResp)

	return strResp, nil
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

	//
	// +++++ ACCESS TO /DATA  +++++++
	//
	//fmt.Println("Private Console")
	//dataPriv, err := accessConsole("GET", PRIVCONSOLE, "DATA", nil, ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 001")
	//	os.Exit(1)
	//}
	//fmt.Println("data from private => ", dataPriv)
	//
	//fmt.Println("Public Console")
	//dataPub, err := accessConsole("GET", PUBCONSOLE, "DATA", nil, ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 002")
	//	os.Exit(1)
	//}
	//fmt.Println("data from private => ", dataPub)

	//
	// CREATE CLAIM MANUALLY VIA SKUPPER
	//
    // Create a claim token manually in west - public"
	//ok := runCmd("skupper", "token", "create", "/tmp/token-west-renato01.yaml", "--expiry", "25m0s", "--name", "renato01", "--password", "rena-senha", "--token-type", "claim", "--uses",  "2", "-n", "west")
	//if !ok {
	//	fmt.Println("Error to create the claim 01 for West")
	//}

	// Create a claim token manually in west - public"
	//ok = runCmd("skupper", "token", "create", "/tmp/token-west-renato02.yaml", "--expiry", "25m0s", "--name", "renato02", "--password", "rena-senha", "--token-type", "claim", "--uses",  "1", "-n", "west")
	//if !ok {
	//	fmt.Println("Error to create the claim 02 for West")
	//}


	//
	// +++++ CREATE CLAIM VIA API  +++++++
	//

	tokenExpires := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
	tokenUses := 3
	postPath := fmt.Sprintf("tokens?expiration=%v&uses=%d", tokenExpires, tokenUses)

	//postPath := "tokens?expiration=60m&uses=3"
	retorno, err := accessConsole("POST", PUBCONSOLE, postPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 003")
		os.Exit(1)
	}
	fmt.Println("retorno ", retorno)

	//// Trying json.RawMessage
	//jsonString = `{"expiration":"", "uses":"2"}`
	//rawJson := json.RawMessage(jsonString)
	//jsonRawReader := bytes.NewReader(rawJson)
	//
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", jsonRawReader, ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 003")
	//	os.Exit(1)
	//}
	//
	//
	//// Try using the opts struct
	//tokenOpts := TokenOptions{
	//	Expiry: 20,
	//	Uses:   4,
	//}
	//jsonBytes, err := json.Marshal(tokenOpts)
	//jsonBytesReader := bytes.NewReader(jsonBytes)
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", jsonBytesReader, ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 004")
	//	os.Exit(1)
	//}
	//
	//// Try using url.values
	//data := url.Values{}
	//data.Set("expiration", "20")
	//data.Add("uses", "20")
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", strings.NewReader(data.Encode()), ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 005")
	//	os.Exit(1)
	//}
	//
	//// Try using map and marshal
	//values := map[string]string{"expiration": "20", "uses": "4"}
	//json_data, err := json.Marshal(values)
	//
	//
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", bytes.NewReader(json_data), ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 005")
	//	os.Exit(1)
	//}
	//
	//// Try using a string as parameter
	//jsonString = `"expiration":"20", "uses":"4"}`
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", strings.NewReader(jsonString), ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 006")
	//	os.Exit(1)
	//}
	//
	//// Try using parameter format
	//jsonString = `expiration=20&uses=4`
	//_, err = accessConsole("POST", PUBCONSOLE, "tokens", strings.NewReader(jsonString), ADMUSER, ADMPASS)
	//if err != nil {
	//	fmt.Println("Erro 007")
	//	os.Exit(1)
	//}

	// Tokens from pub
	tokenPub, err := accessConsole("GET", PUBCONSOLE, "tokens", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 003")
		os.Exit(1)
	}
	fmt.Println("Tokens from pub => ", tokenPub)

	//
	// CREATE A LINK in PRIVATE
	//

	var datafromJson []TokenState
	err = json.Unmarshal([]byte(tokenPub), &datafromJson)
	if err != nil {
		fmt.Println("Error while Unmarshalling data")
	}
	fmt.Println("Data from json")
	fmt.Println(datafromJson)

	fmt.Println("First claim available")
	fmt.Println(datafromJson[0].Name)

	// Tokens from pub

	data := url.Values{}
	data.Set("cost", "20")
	data.Add("token", datafromJson[0].Name)

	linkPriv, err := accessConsole("POST", PRIVCONSOLE, "links", strings.NewReader(data.Encode()), ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 010")
		os.Exit(1)
	}
	fmt.Println("Links from Priv => ", linkPriv)

	//
	// REMOVE TOKENS VIA DELETE
	//
    fmt.Println("Removendo tokens")
	for _, elem := range strings.Split(tokenPub, ",") {
		if strings.Contains(elem, "name") {
			tname := strings.Split(elem, ":")[1]
			tname = strings.Trim(tname, " ")
			name  := strings.Split(tname, "\"")[1]
			fmt.Println("Removing token => ", name)
			_, err := accessConsole("DELETE", PUBCONSOLE, "tokens/" + name, nil, ADMUSER, ADMPASS)
			if err != nil {
				fmt.Println("Unable to remove token => ", name)
			} else {
				fmt.Printf("Token %s removed \n ", name)
			}
		}
	}



}
