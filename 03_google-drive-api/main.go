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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "credentials/token.json"
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
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
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

func getService () *drive.Service {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials/creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	return srv
}

// ========== This section is responsible to fetch files data from a drive folder ==========

func getFolderId (url string) string {
	arr := strings.Split(url, "folders/")

	return arr[1]
}

func getFolderFiles (url string) *drive.FileList{
	folderId := getFolderId(url)
	query := fmt.Sprintf("parents = '%s'", folderId)

	fileList, err := srv.Files.List().Q(query).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	return fileList
}

func getFolderInfos (folderUrl string) []*drive.File {
	fileList := getFolderFiles(folderUrl)
	files := fileList.Files

	nextPageToken := fileList.NextPageToken

	for nextPageToken != "" {
		if len(files) == 0 {
			fmt.Println("No files found.")
			break
		}
		fileList = getFolderFiles(folderUrl)

		files = append(files, fileList.Files...)
	}

	return files
}

var srv *drive.Service = getService()

func main() {
	folderUrl := "https://drive.google.com/drive/u/0/folders/11ftvdwveKCM3HNM0E5SxPNYTlXRyke8L"

	var parents []string
	parents = append(parents, getFolderId(folderUrl))

	folder, err := srv.Files.Create(&drive.File{
		Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		MimeType: "application/vnd.google-apps.folder",
		Parents: parents,
	}).Fields("id").Do()

	if err != nil {
		fmt.Printf("Erro: %s", err)
	}

	files := getFolderInfos(folderUrl)
	
	for index, i := range files {
		fmt.Printf("[%d] %s (%s)\n", index, i.Name, i.Id)
	}

	fmt.Printf("Folder: %#v\n\n aaa: %#v", folder, parents)
}