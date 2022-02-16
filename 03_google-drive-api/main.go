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

// ====================================== Miscelaneous ======================================

func errorPrinter (err error) {
	if err != nil {
		fmt.Printf(`
		------------------------
		The following error occurred:
		%s
		------------------------
		`, err)
	}
}

func prettyPrinter (msgs... string) {
	for _, msg := range(msgs) {
		fmt.Printf(`
	-------------------
	%s
	-------------------`, msg)
	}
}

func getGoDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")
  
	if err != nil {
	  log.Fatalf("Error loading .env file")
	}
  
	return os.Getenv(key)
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

// ========== This section is responsible to create new folders ==========

func checkFolderDuplicates (name string, folderUrl string) bool {
	files := getFolderInfos(folderUrl)

	for _, file := range files {
		if file.Name == name && file.MimeType == "application/vnd.google-apps.folder" {
			return true
		}
	}

	return false
}

func getDuplicate (name string, parentUrl string) *drive.File {
	files := getFolderInfos(parentUrl)

	var file *drive.File = nil

	for _, file = range files {
		if file.Name == name && file.MimeType == "application/vnd.google-apps.folder" {
			break
		}
	}
	
	return file
}

func createFolder (name string, parentUrl string) *drive.File{
	if checkFolderDuplicates(name, parentUrl) {
		return getDuplicate(name, parentUrl)
	}

	var parents []string
	parents = append(parents, getFolderId(parentUrl))

	folder, err := srv.Files.Create(&drive.File{
		Name: name,
		MimeType: "application/vnd.google-apps.folder",
		Parents: parents,
	}).Fields("id").Do()

	errorPrinter(err)
	return folder
}

// ========== This section is responsible for files manipulation ==========

func copyFileTo (file *drive.File, destinationFolder string) {
	var parents []string
	parents = append(parents, destinationFolder)

	file_copied, err := srv.Files.Copy(file.Id, &drive.File{
		Name: file.Name,
		Parents: parents,
	}).Do()

	errorPrinter(err)
	prettyPrinter(fmt.Sprintf("Arquivo copiado:\n%#v", file_copied.Id))
}

var srv *drive.Service = getService()

func main() {
	folderUrl := getGoDotEnvVariable("FOLDER_URL")
	
	newFolder := createFolder("MyNewFolder", folderUrl)
	if newFolder != nil {
		prettyPrinter(fmt.Sprintf("Folder ID: %s", newFolder.Id))
	}
	
	files := getFolderInfos(folderUrl)
	for index, file := range files {
		fmt.Printf(`
		[%d] %s (%s)
		-----------`, index, file.Name, file.Id)

		if strings.Contains(strings.ToLower(file.Name), "grade") {
			copyFileTo(file, newFolder.Id)
		}
	}

}