package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const pricesURL string = "https://www.albion-online-data.com/api/v2/stats/prices"

const credentialsFile string = "credentials.json"
const tokenFile string = "token.json"
const oauth2Scope string = "https://www.googleapis.com/auth/spreadsheets"

const spreadsheetID string = "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
const sheetName string = "MarketData"

type priceData struct {
	prettyName string
	name       string
	sellMin    int
	buyMax     int
}

// Retrieve a sheets service
func getService() (*sheets.Service, error) {
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read credentials.json file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, oauth2Scope)
	if err != nil {
		log.Fatalf("Unable to parse credentials.json to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

/*func getRequestURL() string {

}*/

/*func getPrices() {
	requestURL = getRequestURL(pricesURL)
	resp, err := http.Get()
}*/

func main() {
	srv, err := getService()
	if err != nil {
		log.Fatalf("Cannot get service: %s", err)
	}

	//pricesJson := getPrices()
}
