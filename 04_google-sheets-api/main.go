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

func errorPrinter (err error) {
	if err != nil {
		fmt.Printf(`
		---------------
		Error: %s
		---------------`, err)
	}
}

func prettyPrinter (msg string) {
	fmt.Printf(`
	-----------------
	%s
	-----------------`, msg)
}

func checkSheetDuplicates (spreadsheetUrl string, sheetName string) (bool, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)
	spreadsheet, err := getSpreadsheetInfo(spreadsheetId)
	if err != nil {
		return true, err
	}

	for _, sheet := range(spreadsheet.Sheets) {
		if sheet.Properties.Title == sheetName {
			return true, nil
		}
	}

	return false, err
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
	return readedRange, err
}

func getSpreadsheetInfo (spreadsheetUrl string) (*sheets.Spreadsheet, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	sheet, err := srv.Spreadsheets.Get(spreadsheetId).Do()

	return sheet, err
}


// ============================== Creating Spreadsheet & Tabs ==============================

func createSpreadsheet(spreadsheetTitle string, tabs ...string) (*sheets.Spreadsheet, error) {
	var spreadsheetTabs []*sheets.Sheet
	for _, tabName := range(tabs) {
		spreadsheetTabs = append(spreadsheetTabs, &sheets.Sheet{
			Properties: &sheets.SheetProperties{
				Title: tabName,
			},
		})
	}

	sheet, err := srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: spreadsheetTitle,
		},
		Sheets: spreadsheetTabs,
	}).Do()

	return sheet, err
}

func createNewSheet (spreadsheetUrl string, tabNames ...string) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	var requests []*sheets.Request
	for _, tabName := range(tabNames) {
		isDuplicate, err := checkSheetDuplicates(spreadsheetUrl, tabName)
		if err != nil {
			return nil, err
		}

		if !isDuplicate {
			requests = append(requests, &sheets.Request{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: tabName,
					},
				},
			})
		} else {
			return nil, fmt.Errorf("erro ao criar uma aba com nome %q\njá existe uma aba com este nome", tabName)
		}
	}
	
	update, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
		IncludeSpreadsheetInResponse: true,
	}).Do()
	return update, err
}


// ================================= Updating Spreadsheets =================================

func duplicateSheet (spreadsheetUrl string, sourceSheetId int64, newSheetIndex int64, newSheetName string) (*sheets.BatchUpdateSpreadsheetResponse, error){
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	var requests []*sheets.Request
	isDuplicate, err := checkSheetDuplicates(spreadsheetUrl, newSheetName)
	if err != nil {
		return nil, err
	}

	if !isDuplicate {
		requests = append(requests, &sheets.Request{
			DuplicateSheet: &sheets.DuplicateSheetRequest{
				SourceSheetId: sourceSheetId,
				InsertSheetIndex: newSheetIndex,
				NewSheetName: newSheetName,
			},
		})
	} else {
		return nil, fmt.Errorf("erro ao duplicar a aba com id %d\n Já existe uma aba com o nome %q", sourceSheetId, newSheetName)
	}

	update, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
		IncludeSpreadsheetInResponse: true,
	}).Do()

	return update, err
}

func deleteSheet (spreadsheetUrl string, sheetId int64) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	var requests []*sheets.Request
	requests = append(requests, &sheets.Request{
		DeleteSheet: &sheets.DeleteSheetRequest{
			SheetId: sheetId,
		},
	})

	update, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
		IncludeSpreadsheetInResponse: true,
	}).Do()

	return update, err
}

func updateSpreadsheet (requestedChanges []*sheets.Request, spreadsheetUrl string) (*sheets.BatchUpdateSpreadsheetResponse, error) {
    spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	update, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requestedChanges,
		IncludeSpreadsheetInResponse: true,
	}).Do()

	return update, err
}


// =================================== Writing in a sheet ===================================

func writeSingleRange (spreadsheetUrl string, newLines [][]interface{}, writeRange string) (*sheets.UpdateValuesResponse, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	writedRange, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &sheets.ValueRange{
		Values: newLines,
	}).ValueInputOption("USER_ENTERED").IncludeValuesInResponse(true).Do()

	return writedRange, err
}

func writeMultipleRanges (spreadsheetUrl string, data []*sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	spreadsheetId := getSpreadsheetId(spreadsheetUrl)

	writedRanges, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data: data,
		IncludeValuesInResponse: true,
	}).Do()

	return writedRanges, err
}


// ================================ Sheets service variable ================================

var srv *sheets.Service = getService()


// ============================= Testing the created functions =============================

func main() {
	spreadsheetUrl := getGoDotEnvVariable("GOOGLE_SAMPLE_SPREADSHEET_URL")
	readRange := getGoDotEnvVariable("GOOGLE_SAMPLE_SPREADSHEET_RANGE")
	data, err := getDataFromSpreadsheet(spreadsheetUrl, readRange)
	errorPrinter(err)

	if len(data.Values) == 0 {
		prettyPrinter("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range data.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			prettyPrinter(fmt.Sprintf("%s, %s", row[0], row[4]))
		}
	}


	multipleData, err := getMultipleDataFromSpreadsheet(spreadsheetUrl, readRange)
	errorPrinter(err)
	for rangeNum, vr := range multipleData.ValueRanges {
		fmt.Printf("\nRange: %d", rangeNum)
		for rowNum, rowData := range vr.Values {
			prettyPrinter(fmt.Sprintf("Row: %d - %#v", rowNum, rowData))
		}
	}

	test, err := createSpreadsheet("Será que deu", "Aba 1", "Pedro", "Dev")
	errorPrinter(err)
	prettyPrinter(fmt.Sprintf("Nova aba: %s", test.SpreadsheetUrl))

	mySpreadsheetUrl := getGoDotEnvVariable("MY_SPREADSHEET")

	newSheet, err := createNewSheet(mySpreadsheetUrl, "Mano", "Muito", "brabíssimo")
	if err == nil {
		prettyPrinter("Last tab created: " + newSheet.UpdatedSpreadsheet.Sheets[len(newSheet.UpdatedSpreadsheet.Sheets) - 1].Properties.Title)
	}
	
	sheet, err := getSpreadsheetInfo(mySpreadsheetUrl)
	errorPrinter(err)
	for _, sheetName := range(sheet.Sheets) {
		if sheetName.Properties.Title == "brabíssimo" {
			_, err := duplicateSheet(mySpreadsheetUrl, sheetName.Properties.SheetId, sheetName.Properties.Index + 1, "brabíssimo 2.0")
			errorPrinter(err)
		}
	}

	sheet, err = getSpreadsheetInfo(mySpreadsheetUrl)
	errorPrinter(err)
	var changes []*sheets.Request
	for _, sheetName := range(sheet.Sheets) {
		prettyPrinter(fmt.Sprintf("Deletando: %#v", sheetName.Properties.Title))
		_, err := deleteSheet(mySpreadsheetUrl, sheetName.Properties.SheetId)
		if err != nil {
			changes = append(changes, &sheets.Request{
				DuplicateSheet: &sheets.DuplicateSheetRequest{
					SourceSheetId: sheetName.Properties.SheetId,
					InsertSheetIndex: sheetName.Properties.Index,
					NewSheetName: "New sheet",
				},
			})
		}
	}

	updatedSheet, err := updateSpreadsheet(changes, mySpreadsheetUrl)
	errorPrinter(err)
	if err == nil {
		var multipleWriteData []*sheets.ValueRange
		for _, sheetName := range(updatedSheet.UpdatedSpreadsheet.Sheets) {
			prettyPrinter(fmt.Sprintf("%#v", sheetName.Properties.Title))
	
			_, err := writeSingleRange(mySpreadsheetUrl, data.Values, sheetName.Properties.Title+"!A1")
			errorPrinter(err)
	
			multipleWriteData = append(multipleWriteData, &sheets.ValueRange{
				Range: sheetName.Properties.Title + "!H2",
				Values: data.Values,
			})
			_, err = writeMultipleRanges(mySpreadsheetUrl, multipleWriteData)
			errorPrinter(err)
		}	
	}
}