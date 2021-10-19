package main

import (
	//"bytes"
	"crypto/tls"
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

type LinkStatus struct {
	Name        string
	Url         string
	Cost        int
	Connected   bool
	Configured  bool
	Description string
	Created     string
}



func accessConsole(method string, url string, path string, body io.Reader, user string, pass string) (string, error) {

	// Define the request first
	req, err := http.NewRequest(method, url+"/"+path, body)
	if err != nil {
		return "", fmt.Errorf("Error defining the request")
	}

	// If this is a POST
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		//req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
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

	fmt.Println("Resp Body => ", strResp)
	fmt.Println("Resp Header => ", resp.Header)

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

	//
	// +++++ CREATE CLAIM VIA API  +++++++
	//

	tokenExpires := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
	tokenUses := 3
	postPath := fmt.Sprintf("tokens?expiration=%v&uses=%d", tokenExpires, tokenUses)

	//postPath := "tokens?expiration=60m&uses=3"
	tokenCreatedPublic, err := accessConsole("POST", PUBCONSOLE, postPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 003")
		os.Exit(1)
	}
	fmt.Println("Token Created in Public", tokenCreatedPublic)

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
	data := url.Values{}
	data.Set("cost", "20")
	if err != nil {
		fmt.Println("Erro 010")
		os.Exit(1)
	}
	_, err = accessConsole("POST", PRIVCONSOLE, "links?cost=20", strings.NewReader(tokenCreatedPublic), ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 010")
		os.Exit(1)
	}

	// List links from pub
	linkPriv, err := accessConsole("GET", PRIVCONSOLE, "links", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 003")
		os.Exit(1)
	}
	fmt.Println("Links from Priv => ", linkPriv)

	// Tokens from pub
	tokenPub, err = accessConsole("GET", PUBCONSOLE, "tokens", nil, ADMUSER, ADMPASS)
	if err != nil {
		fmt.Println("Erro 003")
		os.Exit(1)
	}
	fmt.Println("Tokens from pub => ", tokenPub)


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
