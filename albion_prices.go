package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const credentialsFilePath string = "credentials.json"
const tokenFilePath string = "token.json"
const oauth2Scope string = "https://www.googleapis.com/auth/spreadsheets"

const spreadsheetID string = ""
const sheetName string = "MarketData"
const sheetPosition string = "A1"

const pricesURL string = "https://www.albion-online-data.com/api/v2/stats/prices"
const pricesRequestItemsLenCap int = 200

const defaultItemDataDumpFilePath string = "items.txt"
const defaultEnchantableResourcesFilePath string = "enchantableResources.txt"
const defaultUnenenchantableItemsFilePath string = "unenchantableItems.txt"

var enchResSuffixes = [3]string{"_LEVEL1@1", "_LEVEL2@2", "_LEVEL3@3"}

func findStringID(slice []string, key string) int {
	for id := range slice {
		if slice[id] == key {
			return id
		}
	}
	return -1
}

// Retrieve a sheets service
func getService() (*sheets.Service, error) {
	b, err := ioutil.ReadFile(credentialsFilePath)
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
	tok, err := tokenFromFile(tokenFilePath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFilePath, tok)
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
	fmt.Printf("Saving oauth token file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func feedItemNames(out chan<- string) {
	enchResNames := genEnchantableResourceNames(defaultEnchantableResourcesFilePath)
	for _, name := range enchResNames {
		out <- name
	}
	unenchItemNames := getUnenchantableItemNames(defaultUnenenchantableItemsFilePath)
	for _, name := range unenchItemNames {
		out <- name
	}

	close(out)
}

func genEnchantableResourceNames(enchantableResourcesFilePath string) []string {
	var enchResNames []string = make([]string, 0)

	f, err := os.Open(enchantableResourcesFilePath)
	if err != nil {
		log.Fatalf("Couldn't open enchantable resource names file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		resource := scanner.Text()
		resEnch1 := resource + enchResSuffixes[0]
		resEnch2 := resource + enchResSuffixes[1]
		resEnch3 := resource + enchResSuffixes[2]
		enchResNames = append(enchResNames, resource, resEnch1, resEnch2, resEnch3)
	}

	err = scanner.Err()
	if err != nil {
		log.Fatalf("Error reading enchantable resource names file: %s", err)
	}

	return enchResNames
}

func getUnenchantableItemNames(unenchantableItemsFilePath string) []string {
	var itemNames []string = make([]string, 0)

	f, err := os.Open(unenchantableItemsFilePath)
	if err != nil {
		log.Fatalf("Couldn't open unenchantable items names file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		item := scanner.Text()
		itemNames = append(itemNames, item)
	}

	err = scanner.Err()
	if err != nil {
		log.Fatalf("Error reading unenchantable items names file: %s", err)
	}

	return itemNames
}

func getPrices(items <-chan string, responses chan<- []byte) {
	var itemsBatch []string = make([]string, 0)
	var batchLen int = 0

	for itm := range items {
		if (batchLen + len(itm)) > pricesRequestItemsLenCap {
			fmt.Printf("Requesting %s\n", strings.Join(itemsBatch, " "))
			getPricesBatch(itemsBatch, responses)
			itemsBatch = itemsBatch[:0]
			batchLen = 0
		}
		itemsBatch = append(itemsBatch, itm)
		batchLen += len(itm)
	}

	if batchLen != 0 {
		getPricesBatch(itemsBatch, responses)
	}
	close(responses)
}

func getPricesBatch(batch []string, responses chan<- []byte) {
	requestURL := getRequestURL(batch)

	resp, err := http.Get(requestURL)
	if err != nil {
		log.Fatalf("HTTP GET %s failed:", requestURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Reading HTTP GET %s's body failed:", requestURL, err)
	}

	responses <- body
}

func getRequestURL(items []string) string {
	var b strings.Builder
	b.WriteString(pricesURL)
	b.WriteRune('/')

	for _, itm := range items {
		b.WriteString(itm)
		b.WriteRune(',')
	}

	return b.String()
}

func genSheetValues(marketData <-chan []byte) sheets.ValueRange {
	var vr sheets.ValueRange
	var cities []string = make([]string, 0)
	var itemToRow map[string]int = make(map[string]int)
	var rowIter int = 1

	// The city row
	vr.Values = append(vr.Values, []interface{}{""})

	for response := range marketData {
		var itms []map[string]interface{}
		json.Unmarshal(response, &itms)

		for _, itm := range itms {
			itemID := itm["item_id"].(string)
			city := itm["city"].(string)
			sellMin := int(itm["sell_price_min"].(float64))
			buyMax := int(itm["buy_price_max"].(float64))

			itemRowNum, ok := itemToRow[itemID]
			if !ok {
				itemToRow[itemID] = rowIter
				itemRowNum = rowIter

				newValueRow := []interface{}{itemID}
				vr.Values = append(vr.Values, newValueRow)
				rowIter++

				if len(vr.Values) != rowIter {
					panic("coding mistake in genSheetValues")
				}
			}

			cityID := findStringID(cities, city)
			if cityID == -1 {
				// new city, extend cities and city row
				cityID = len(cities)
				cities = append(cities, city)
				vr.Values[0] = append(vr.Values[0], city, city)
			}
			cityColumn := 1 + (cityID * 2) // each city takes 2 columns (sell min, buy max) and 1st column is empty

			if extendColumns := (cityColumn + 2) - len(vr.Values[itemRowNum]); extendColumns > 0 {
				// Item's row is too short, extend it
				vr.Values[itemRowNum] = append(vr.Values[itemRowNum], make([]interface{}, extendColumns)...)
			}

			vr.Values[itemRowNum][cityColumn+0] = sellMin
			vr.Values[itemRowNum][cityColumn+1] = buyMax
		}
	}

	return vr
}

func getSheetRange() string {
	return fmt.Sprintf("%s!%s", sheetName, sheetPosition)
}

func main() {
	srv, err := getService()
	if err != nil {
		log.Fatalf("Cannot get service: %s", err)
	}

	itemCh := make(chan string, 32)
	respCh := make(chan []byte, 16)
	go feedItemNames(itemCh)
	go getPrices(itemCh, respCh)

	vr := genSheetValues(respCh)

	srv.Spreadsheets.Values.Update(spreadsheetID, getSheetRange(), &vr).ValueInputOption("USER_ENTERED").Do()
}
