package main

import (
	"bytes"
	//"bytes"
	"crypto/tls"
	"encoding/json"
	//"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
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

type LinkStatus struct {
	Name        string
	Url         string
	Cost        int
	Connected   bool
	Configured  bool
	Description string
	Created     string
}

type ServiceEndpoint struct {
	Name   string      `json:"name"`
	Target string      `json:"target"`
	Ports  map[int]int `json:"ports,omitempty"`
}

type ServiceDefinition struct {
	Name      string            `json:"name"`
	Protocol  string            `json:"protocol"`
	Ports     []int             `json:"ports"`
	Endpoints []ServiceEndpoint `json:"endpoints"`
}

type PortDescription struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type ServiceTarget struct {
	Name  string            `json:"name"`
	Type  string            `json:"type"`
	Ports []PortDescription `json:"ports,omitempty"`
}

type ServiceOptions struct {
	Address     string            `json:"address"`
	Protocol    string            `json:"protocol"`
	Ports       []int             `json:"ports"`
	TargetPorts map[int]int       `json:"targetPorts,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Target      ServiceTarget     `json:"target"`
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

	//fmt.Println("Resp Body => ", strResp)
	//fmt.Println("Resp Header => ", resp.Header)
	//fmt.Println("Resp Status => ", resp.StatusCode)

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

// Check if a linkStatus structure contains an element based on its name
func findInSlice(elements []LinkStatus, key string) bool {

	for _, element := range elements {
		if element.Name == key {
			return true
		}
	}
	return false
}

// Return the first element in slice1 not found in slice2
func findNewLink(elementsAfter []LinkStatus, elementsBefore []LinkStatus) (string, error) {
	founds := 0
	newLink := ""
	for _, nameAfter := range elementsAfter {
		if findInSlice(elementsBefore, nameAfter.Name) == false {
			newLink = nameAfter.Name
			founds++
		}
	}
	if founds == 1 {
		return newLink, nil
	} else {
		return "", fmt.Errorf("More than 1 newLink found")
	}
}

func testAccessDATA(consoleUrl string) (string, error) {

	dataConsole, err := accessConsole("GET", consoleUrl, "DATA", nil, ADMUSER, ADMPASS)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve /DATA from %s", consoleUrl)
	}
	return dataConsole, nil
}

//  TOKENS FUNCTIONS
func getTokens(consoleUrl string) ([]TokenState, error) {

	tokensCreatedStr, err := accessConsole("GET", consoleUrl, "tokens", nil, ADMUSER, ADMPASS)
	if err != nil {
		return []TokenState{}, fmt.Errorf("Unable to list tokens from %s", consoleUrl)
	}

	var tokensCreated []TokenState
	err = json.Unmarshal([]byte(tokensCreatedStr), &tokensCreated)
	if err != nil {
		return []TokenState{}, fmt.Errorf("Unable to unmarshal tokens list for %s", consoleUrl)
	}
	return tokensCreated, nil
}

func getOneToken(consoleUrl string, claimID string) (TokenState, error) {

	getPath := fmt.Sprintf("tokens/%s", claimID)

	tokenGotStr, err := accessConsole("GET", consoleUrl, getPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return TokenState{}, fmt.Errorf("Unable to retrieve claim %s from %s", claimID, consoleUrl)
	}

	var tokenGot TokenState
	err = json.Unmarshal([]byte(tokenGotStr), &tokenGot)
	if err != nil {
		return TokenState{}, fmt.Errorf("Unable to unmarshal retrieved claim %s", claimID)
	}
	return tokenGot, nil
}

func downloadClaimToken(consoleUrl string, claimID string) (corev1.Secret, error) {

	postPath := fmt.Sprintf("downloadclaim/%s", claimID)

	tokenDownloadedStr, err := accessConsole("GET", consoleUrl, postPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return corev1.Secret{}, fmt.Errorf("Unable to download claim %s from %s", claimID, consoleUrl)
	}

	var tokenDownloaded corev1.Secret
	err = json.Unmarshal([]byte(tokenDownloadedStr), &tokenDownloaded)
	if err != nil {
		return corev1.Secret{}, fmt.Errorf("Unable to unmarshal downloaded claim %s", claimID)
	}
	return tokenDownloaded, nil
}

func createClaimToken(consoleUrl string, minutes int, uses int) (corev1.Secret, error) {

	tokenExpires := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
	postPath := fmt.Sprintf("tokens?expiration=%v&uses=%d", tokenExpires, uses)

	tokenCreatedStr, err := accessConsole("POST", consoleUrl, postPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return corev1.Secret{}, fmt.Errorf("Unable to create token for %s", consoleUrl)
	}

	var tokenCreated corev1.Secret
	err = json.Unmarshal([]byte(tokenCreatedStr), &tokenCreated)
	//fmt.Println("Debug Token Created = ", tokenCreatedStr)
	if err != nil {
		return corev1.Secret{}, fmt.Errorf("Unable to unmarshal token for %s", consoleUrl)
	}
	return tokenCreated, nil
}

func delToken(consoleUrl string, tokenName string) (error) {
	_, err := accessConsole("DELETE", consoleUrl, "tokens/" + tokenName, nil, ADMUSER, ADMPASS)
	if err != nil {
		return fmt.Errorf("Unable to delete claim token %s from %s", tokenName, consoleUrl)
	}
	return nil
}

func printClaim(claim TokenState) {
	fmt.Printf("\nNAME => %s\n", claim.Name)
	fmt.Println("  EXPIRY => ", *claim.ClaimExpiration)
	if claim.ClaimsMade == nil {
		fmt.Println("  CLAIMS MADE => ", 0)
	} else {
		fmt.Println("  CLAIMS MADE => ", *claim.ClaimsMade)
	}
	fmt.Println("  CLAIMS REMAINING => ", *claim.ClaimsRemaining)
}


//  LINKS FUNCTIONS
func getLinks(consoleUrl string) ([]LinkStatus, error) {

	linksCreatedSTR, err := accessConsole("GET", consoleUrl, "links", nil, ADMUSER, ADMPASS)
	if err != nil {
		return []LinkStatus{}, fmt.Errorf("Unable to list links from %s", consoleUrl)
	}

	var linksCreated []LinkStatus
	err = json.Unmarshal([]byte(linksCreatedSTR), &linksCreated)
	if err != nil {
		return []LinkStatus{}, fmt.Errorf("Unable to unmarshal link list for %s", consoleUrl)
	}
	return linksCreated, nil
}

func getOneLink(consoleUrl string, linkID string) (LinkStatus, error) {

	getPath := fmt.Sprintf("links/%s", linkID)

	linkGotStr, err := accessConsole("GET", consoleUrl, getPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return LinkStatus{}, fmt.Errorf("Unable to retrieve link %s from %s", linkID, consoleUrl)
	}

	var linkGot LinkStatus
	err = json.Unmarshal([]byte(linkGotStr), &linkGot)
	if err != nil {
		return LinkStatus{}, fmt.Errorf("Unable to unmarshal retrieved link %s", linkID)
	}
	return linkGot, nil
}

func createLink(consoleUrl string, cost int, secret corev1.Secret ) error {

	postPath := fmt.Sprintf("links?cost=%d", cost)
	secretSTR, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal token for %s", consoleUrl)
	}

	_, err = accessConsole("POST", consoleUrl, postPath, bytes.NewReader(secretSTR), ADMUSER, ADMPASS)
	if err != nil {
		return fmt.Errorf("Unable to create token for %s", consoleUrl)
	}
	return nil
}

func delLink(consoleUrl string, linkName string) (error) {
	_, err := accessConsole("DELETE", consoleUrl, "links/" + linkName, nil, ADMUSER, ADMPASS)
	if err != nil {
		return fmt.Errorf("Unable to delete link %s from %s", linkName, consoleUrl)
	}
	return nil
}

func lastSlice(fullString string, sep string) string {
	slicedString := strings.Split(fullString, sep)
    return string(slicedString[len(slicedString)-1])
}

func printLink(link LinkStatus) {
	fmt.Printf("\nNAME => %s\n", link.Name)
	fmt.Println("  URL => ", link.Url)
	fmt.Println("  COST => ", link.Cost)
	fmt.Println("  CONFIGURED => ", link.Configured)
	fmt.Println("  CONNECTED => ", link.Connected)
	fmt.Println("  DESCRIPTION => ", link.Description)
	fmt.Println("  CREATED => ", link.Created)
}


//  SERVICES FUNCTIONS
func getServices(consoleUrl string) ([]ServiceDefinition, error) {

	servicesStr, err := accessConsole("GET", consoleUrl, "services", nil, ADMUSER, ADMPASS)
	if err != nil {
		return []ServiceDefinition{}, fmt.Errorf("Unable to list services from %s", consoleUrl)
	}

	var services []ServiceDefinition
	err = json.Unmarshal([]byte(servicesStr), &services)
	if err != nil {
		return []ServiceDefinition{}, fmt.Errorf("Unable to unmarshal service list for %s", consoleUrl)
	}
	return services, nil
}

func createService(consoleUrl string, service ServiceOptions) error {

	postPath := "services"
	serviceSTR, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal service for %s", consoleUrl)
	}

	_, err = accessConsole("POST", consoleUrl, postPath, bytes.NewReader(serviceSTR), ADMUSER, ADMPASS)
	if err != nil {
		return fmt.Errorf("Unable to create service for %s", consoleUrl)
	}
	return nil
}

func getOneService(consoleUrl string, serviceID string) (ServiceDefinition, error) {

	getPath := fmt.Sprintf("services/%s", serviceID)

	serviceStr, err := accessConsole("GET", consoleUrl, getPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return ServiceDefinition{}, fmt.Errorf("Unable to retrieve service %s", serviceID)
	}

	var service ServiceDefinition
	err = json.Unmarshal([]byte(serviceStr), &service)
	if err != nil {
		return ServiceDefinition{}, fmt.Errorf("Unable to unmarshal service %s", serviceID)
	}
	return service, nil
}

func delService(consoleUrl string, serviceName string) (error) {

	_, err := accessConsole("DELETE", consoleUrl, "services/" + serviceName, nil, ADMUSER, ADMPASS)
	if err != nil {
		return fmt.Errorf("Unable to delete service %s from %s", serviceName, consoleUrl)
	}
	return nil
}

func printService(svc ServiceDefinition) {
	fmt.Printf("\nNAME => %s\n", svc.Name)
	fmt.Println("  PROTOCOL => ", svc.Protocol)

	for _, port := range svc.Ports {
		fmt.Println("  PORTS => ", port)
	}

	for _, endpoint := range svc.Endpoints {
		fmt.Println("  ENDPOINT NAME => ", endpoint.Name)
		fmt.Println("  ENDPOINT TARGET => ", endpoint.Target)
		for _, port := range endpoint.Ports {
			fmt.Println("  ENDPOINT PORTS => ", port)
		}
	}
}

//  TARGET FUNCTIONS
func getTargets(consoleUrl string) ([]ServiceTarget, error) {

	getPath := fmt.Sprintf("targets")

	targetsStr, err := accessConsole("GET", consoleUrl, getPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return []ServiceTarget{}, fmt.Errorf("Unable to list targets from %s", consoleUrl)
	}

	var targets []ServiceTarget
	err = json.Unmarshal([]byte(targetsStr), &targets)
	if err != nil {
		return []ServiceTarget{}, fmt.Errorf("Unable to unmarshal targets from %s", consoleUrl)
	}
	return targets, nil
}

func printTargets(target ServiceTarget) {

	fmt.Printf("\nNAME => %s\n", target.Name)
	fmt.Println("  TYPE => ", target.Type)

	for _, port := range target.Ports {
		fmt.Println("  PORT NAME => ", port.Name)
		fmt.Println("  |--- PORT => ", port.Port)
	}
}


// General Functions
func getGenericEndpoint(consoleUrl string, getPath string) (string, error) {

	genEndPointStr, err := accessConsole("GET", consoleUrl, getPath, nil, ADMUSER, ADMPASS)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve endpoint %s", getPath)
	}
	return genEndPointStr, nil
}



func main() {

	fmt.Println("Here we go !!")

	//
	// +++++ /DATA
	//
	_, err := testAccessDATA(PUBCONSOLE)
	if err != nil {
		fmt.Println("Unable to access /DATA from pubConsole")
	}

	_, err = testAccessDATA(PRIVCONSOLE)
	if err != nil {
		fmt.Println("Unable to access /DATA from privConsole")
	}

	//
	// +++++ LIST TOKENS IN PUB
	//
	fmt.Printf("\nListing Claims Tokens in PUB\n============================================\n")
	tokensInPub, err := getTokens(PUBCONSOLE)
	for _, token := range tokensInPub {
		printClaim(token)
	}

	//
	// +++++ CREATE CLAIM TOKEN VIA API
	//
	fmt.Printf("\nCreating a Claim Tokens in PUB via API\n============================================\n")
	pubClaimCreated, err := createClaimToken(PUBCONSOLE, 5, 2)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Claim created with ID %s\n", pubClaimCreated.Name)
	fmt.Printf("Claim created with URL %s\n", pubClaimCreated.Annotations["skupper.io/url"])

	//
	// +++++ DOWNLOAD A CLAIM
	//
	fmt.Printf("\nDownloading a Claim Tokens from PUB via API\n============================================\n")
	claimToDownload := lastSlice(pubClaimCreated.Annotations["skupper.io/url"], "/")
	pubClaimDownloaded, err := downloadClaimToken(PUBCONSOLE, claimToDownload)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Claim Downloaded with ID %s\n", pubClaimDownloaded.Name)
	fmt.Printf("Claim downloaded with URL %s\n", pubClaimDownloaded.Annotations["skupper.io/url"])

	//
	// +++++ LIST TOKENS IN PUB
	//
	fmt.Printf("\nListing Claim Tokens in PUB after creation\n============================================\n")
	tokensInPub, err = getTokens(PUBCONSOLE)
	for _, token := range tokensInPub {
		printClaim(token)
	}

	//
	// +++++ LIST LINKS IN PRIV
	//
	fmt.Printf("\nListing links in PRIV before first link creation\n============================================\n")
	linksInPrivBefore, err := getLinks(PRIVCONSOLE)
	for _, link := range linksInPrivBefore {
		printLink(link)
	}

	//
	// +++++ CREATE LINK IN PRIVATE
	//
	err = createLink(PRIVCONSOLE, 4, pubClaimCreated )
	if err != nil {
		fmt.Println(err)
	}
	// Wait until link get established
	time.Sleep(30 * time.Second)

	//
	// +++++ LIST LINKS IN PRIV AFTER LINK CREATION
	//
	fmt.Printf("\nListing links in PRIV after first link creation\n============================================\n")
	linksInPrivAfter, err := getLinks(PRIVCONSOLE)
	for _, link := range linksInPrivAfter {
		printLink(link)
	}

	newLink, err  := findNewLink(linksInPrivAfter, linksInPrivBefore)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The first link created is %s\n", newLink)

	//
	// +++++ RETRIEVE ONE SPECIFIC LINK
	//
	newLinkData, err := getOneLink(PRIVCONSOLE, newLink)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nDetails about the first link\n============================================\n")
	printLink(newLinkData)

	//
	// +++++ GET CLAIM TOKENS USED IN LINK AND CHECK ITS USES
	//
	claimToGet := lastSlice(pubClaimCreated.Annotations["skupper.io/url"], "/")
	retrievedClaim, err := getOneToken(PUBCONSOLE, claimToGet)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nThis is the token after we used it 1 time")
	printClaim(retrievedClaim)

	//
	// +++++ CREATE A SECOND LINK IN PRIVATE
	//
	err = createLink(PRIVCONSOLE, 3, pubClaimDownloaded )
	if err != nil {
		fmt.Println(err)
	}
	// Wait until link get established
	time.Sleep(30 * time.Second)

	//
	// +++++ LIST LINKS IN PRIV AFTER SECOND LINK CREATION
	//
	linksInPrivBefore = linksInPrivAfter
	fmt.Printf("\nListing links in PRIV after second link creation\n============================================\n")
	linksInPrivAfter, err = getLinks(PRIVCONSOLE)
	for _, link := range linksInPrivAfter {
		printLink(link)
	}

	newLink, err  = findNewLink(linksInPrivAfter, linksInPrivBefore)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The second link created is %s\n", newLink)
	newLinkData, err = getOneLink(PRIVCONSOLE, newLink)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nDetails about the second link\n============================================\n")
	printLink(newLinkData)

	//
	// +++++ GET CLAIM TOKENS USED IN LINK AND CHECK ITS USES
	//
	claimToGet = lastSlice(pubClaimDownloaded.Annotations["skupper.io/url"], "/")
	retrievedClaim, err = getOneToken(PUBCONSOLE, claimToGet)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nThis is the token after we used it 2 times\n============================================\n")
	printClaim(retrievedClaim)

	//
	// +++++ TRY TO CREATE A THIRD LINK IN PRIVATE, IT MUST FAIL
	// +++++ BECAUSE THERE ARE NO AVAILABLE CLAIMS
	//
	linksInPrivBefore = linksInPrivAfter
	err = createLink(PRIVCONSOLE, 2, pubClaimDownloaded )
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nListing links in PRIV after third link creation\n============================================\n")
	linksInPrivAfter, err = getLinks(PRIVCONSOLE)
	for _, link := range linksInPrivAfter {
		printLink(link)
	}
	newLink, err  = findNewLink(linksInPrivAfter, linksInPrivBefore)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The third link created is %s\n", newLink)

	newLinkData, err = getOneLink(PRIVCONSOLE, newLink)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nDetails about the third link")
	printLink(newLinkData)

	//
	// +++++ CREATE A SERVICE IN PRIV
	//
	newsvc := ServiceOptions{
		Address:     "hello-world-backend",
		Protocol:    "http",
		Ports:       []int{8080},
		TargetPorts: map[int]int{8080: 8080},
		Labels:      nil,
		Target:      ServiceTarget{
			Name:  "hello-world-backend",
			Type:  "deployment",
			Ports: []PortDescription{
				 { Name: "8080",
				   Port: 8080},
			},
		},
	}
	fmt.Printf("\nCreating service hello-world-backend in PRIV\n============================================\n")
	err = createService(PRIVCONSOLE, newsvc)
	if err != nil {
		fmt.Println("Unable to create service ", err)
	}

	//
	// +++++ List Services from Pub
	//
	fmt.Printf("\nListing Services in PUB\n============================================\n")
	svcsPub, err := getServices(PUBCONSOLE)
	if err != nil {
		fmt.Println(err)
	}
	for _, svcPub := range svcsPub {
		if svcPub.Endpoints == nil {
			fmt.Println("Service Exposed through Skupper")
		} else {
			fmt.Println("Service Exposed BY Skupper")
		}
		printService(svcPub)
	}

	//
	// +++++ List Services from Priv
	//
	fmt.Printf("\nListing Services in PRIV\n============================================\n")
	svcsPriv, err := getServices(PRIVCONSOLE)
	if err != nil {
		fmt.Println(err)
	}
	for _, svcPriv := range svcsPriv {
		if svcPriv.Endpoints == nil {
			fmt.Println("Service Exposed through Skupper")
		} else {
			fmt.Println("Service Exposed BY Skupper")
		}
		printService(svcPriv)
	}

	//
	// +++++ RETRIEVE ONE SPECIFIC SERVICE
	//
	fmt.Printf("\nRetrieve one specific service in PRIV\n============================================\n")
	oneService, err := getOneService(PRIVCONSOLE, "hello-world-backend")
	if err != nil {
		fmt.Println(err)
	}
	if oneService.Endpoints == nil {
		fmt.Println("Service Exposed through Skupper")
	} else {
		fmt.Println("Service Exposed BY Skupper")
	}
	printService(oneService)

	//
	// +++++ LIST TARGETS FROM A SERVICE
	//
	fmt.Printf("\nListing Targets in PRIV\n============================================\n")
	targetsInSvc, err := getTargets(PRIVCONSOLE)
	if err != nil {
		fmt.Println(err)
	}
	for _, target := range targetsInSvc {
		printTargets(target)
	}

	//
	// +++++ SERVICECHECK
	//
	fmt.Printf("\nChecking Service hello-world-backend\n============================================\n")
	chkSvc, err := getGenericEndpoint(PRIVCONSOLE, "servicecheck/hello-world-backend")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(chkSvc)

	//
	// +++++ GET VERSION
	//
	fmt.Printf("\nGet Versions\n============================================\n")
	version, err := getGenericEndpoint(PUBCONSOLE, "version")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(version)

	//
	// +++++ GET SITE
	//
	fmt.Printf("\nGet Site\n============================================\n")
	site, err := getGenericEndpoint(PUBCONSOLE, "site")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(site)

	//
	// +++++ GET EVENTS
	//
	fmt.Printf("\nGet Events\n============================================\n")
	events, err := getGenericEndpoint(PUBCONSOLE, "events")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(events)

	//
	// +++++ REMOVE SERVICE
	//
	fmt.Printf("\nRemoving Targets from PRIV\n============================================\n")
	svcsInPriv, err := getServices(PRIVCONSOLE)
	for _, svc := range svcsInPriv {
		err := delService(PRIVCONSOLE, svc.Name)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf(  "Service %s removed from PRIV\n", svc.Name)
		}
	}

	//
	// +++++ REMOVE LINKS FROM PRIV
	//
	fmt.Printf("\nRemoving links from PRIV\n============================================\n")
	linksInPriv, err := getLinks(PRIVCONSOLE)
	for _, link := range linksInPriv {
		err := delLink(PRIVCONSOLE, link.Name)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf(  "Link %s removed from PRIV\n", link.Name)
		}
	}

	//
	// +++++ REMOVE CLAIM TOKENS FROM PUB
	//
	fmt.Printf("\nRemoving Claim Tokens from PUB\n============================================\n")
	tokensInPub, err = getTokens(PUBCONSOLE)
	for _, token := range tokensInPub {
		err := delToken(PUBCONSOLE, token.Name)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Claim token %s removed from Pub\n",token.Name)
		}
	}
}
