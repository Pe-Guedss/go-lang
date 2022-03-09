package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// ================================= Client Authentication =================================

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "./credentials/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
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

// Gets the service used to make every drive operation
func getService () *sheets.Service {
	ctx := context.Background()
	b, err := ioutil.ReadFile("./credentials/creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}


// =============================== General Purpose Functions ===============================

// Retrieves environment variables from the ".env" file stored in this repo.
func getGoDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")
  
	if err != nil {
	  log.Fatalf("Error loading .env file")
	}
  
	return os.Getenv(key)
}

func getSpreadsheetId (url string) string {
	if strings.HasPrefix(url, "https") {
		arr := strings.Split(url, "/")
        return arr[5]
	} else {
		return url
	}
}


// ===================================== Reading sheets =====================================

func getDataFromSpreadsheet (spreadsheetUrl string, readRange string) (readedRange *sheets.ValueRange, err error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	readedRange, err = srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	return readedRange, err
}

func getMultipleDataFromSpreadsheet (spreadsheetUrl string, readRange ...string) (readedRange *sheets.BatchGetValuesResponse, err error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	readedRange, err = srv.Spreadsheets.Values.BatchGet(spreadsheetId).Ranges(readRange...).Do()
	// Get(spreadsheetId, readRange).Do()
	return readedRange, err
}


// ================================ Sheets service variable ================================

var srv *sheets.Service = getService()


// ============================= Testing the created functions =============================

func main() {
	_ = getGoDotEnvVariable("TEST")

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
	readRange := "Class Data!A2:E"
	data, err := getDataFromSpreadsheet(spreadsheetId, readRange)
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(data.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range data.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[4])
		}
	}


	multipleData, err := getMultipleDataFromSpreadsheet(spreadsheetId, readRange)
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	for rangeNum, vr := range multipleData.ValueRanges {
		fmt.Printf("Range: %d\n", rangeNum)
		for j, v := range vr.Values {
			fmt.Printf("Index: %d - %#v\n", j, v)
		}
	}
}