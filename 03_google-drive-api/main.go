package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func checkFileDuplicates (currentFile *drive.File, folderUrl string) (bool, error) {
	files, err := getFolderInfos(folderUrl)

	for _, file := range files {
		if file.Name == currentFile.Name && 
		   file.MimeType == currentFile.MimeType {
			return true, err
		}
	}

	return false, err
}

func getDuplicate (currentFile *drive.File, parentUrl string) (*drive.File, error) {
	files, err := getFolderInfos(parentUrl)

	var file *drive.File = nil

	for _, file = range files {
		if file.Name == currentFile.Name && 
		   file.MimeType == currentFile.MimeType {
			break
		}
	}
	
	return file, err
}

// ========== This section is responsible to fetch files data from a drive folder ==========

func getFolderId (url string) string {
	if strings.HasPrefix(url, "https"){
		arr := strings.Split(url, "folders/")

		return arr[1]
	}

	return url
}

func getFolderFiles (url string) (*drive.FileList, error){
	folderId := getFolderId(url)
	query := fmt.Sprintf("parents = '%s'", folderId)

	fileList, err := srv.Files.List().Q(query).Do()

	return fileList, err
}

func getFolderInfos (folderUrl string) ([]*drive.File, error) {
	fileList, err := getFolderFiles(folderUrl)
	files := fileList.Files

	nextPageToken := fileList.NextPageToken

	for nextPageToken != "" {
		if len(files) == 0 {
			fmt.Println("No files found.")
			break
		}
		fileList, err = getFolderFiles(folderUrl)

		files = append(files, fileList.Files...)
	}

	return files, err
}

// ========== This section is responsible to create new folders ==========

func createFolder (name string, parentUrl string) (*drive.File, error){
	var parents []string
	parents = append(parents, getFolderId(parentUrl))

	newFolder := &drive.File{
		Name: name,
		MimeType: "application/vnd.google-apps.folder",
		Parents: parents,
	}

	isDuplicate, err := checkFileDuplicates(newFolder, parentUrl)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		prettyPrinter("This folder already exists!")
		return getDuplicate(newFolder, parentUrl)
	}

	folder, err := srv.Files.Create(newFolder).Fields("id").Do()
	return folder, err
}

// ========== This section is responsible for files manipulation ==========

func copyFileTo (file *drive.File, destinationFolderId string) (*drive.File, error) {

	isDuplicate, err := checkFileDuplicates(file, destinationFolderId)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		return getDuplicate(file, destinationFolderId)
	}

	var parents []string
	parents = append(parents, destinationFolderId)

	fileCopied, err := srv.Files.Copy(file.Id, &drive.File{
		Name: file.Name,
		Parents: parents,
	}).Do()

	errorPrinter(err)
	return fileCopied, err
}

func createFileInsideOf (file *drive.File) (*drive.File, error) {
	for _, parentId := range(file.Parents){
		isDuplicate, err := checkFileDuplicates(file, parentId)
		if err != nil {
			return nil, err
		}
		if isDuplicate {
			prettyPrinter(fmt.Sprintf("There is already a file with this name and type inside the destination folder.\nThe folder ID is: %s", parentId))
			return getDuplicate(file, parentId)
		}
	}

	fileCreated, err := srv.Files.Create(file).Fields("id").Do()

	return fileCreated, err
}

func moveFileTo (source string, target string, file *drive.File) (*drive.File, error) {
	sourceId := getFolderId(source)
	targetId := getFolderId(target)

	movedFile, err := srv.Files.Update(
		file.Id,
		&drive.File{},
	).AddParents(targetId).RemoveParents(sourceId).Do()

	return movedFile, err
}

func uploadFiles (file *os.File, targetDriveFolder string) (*drive.File, error) {
	targetDriveFolder = getFolderId(targetDriveFolder)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	uploadedfile, err := srv.Files.Create(&drive.File{
		Name: fileInfo.Name(),
		Parents: []string{targetDriveFolder},
	}).Media(file).Do()

	return uploadedfile, err
}

func downloadFiles (file *drive.File, localPath string, fileFormat string) error{
	data, err := srv.Files.Get(file.Id).Download()
	if data != nil{
		os.Chdir(localPath)
		dowloadedFile, err := os.Create(fmt.Sprintf("%s.%s", file.Name, fileFormat))
		if err != nil {
			return err
		}
		io.Copy(dowloadedFile, data.Body)
	}

	return err
}

func permanentlyDeleteFile (fileId string) error {
	err := srv.Files.Delete(fileId).Do()
	return err
}

func emptyTrash () (error) {
	err := srv.Files.EmptyTrash().Do()
	return err
}

// ============================= Variável de Serviço do Drive =============================

var srv *drive.Service = getService()

// ============================== Chamada das funções criadas ==============================

func main() {
	parentFolderUrl := getGoDotEnvVariable("PARENT_FOLDER_URL")
	
	newFolder, err := createFolder("MyNewFolder", parentFolderUrl)
	errorPrinter(err)
	if newFolder != nil {
		prettyPrinter(fmt.Sprintf("Folder ID: %s", newFolder.Id))
	}

	createdFile, err := createFileInsideOf(&drive.File{
		Name: "Meu Arquivo",
		MimeType: "application/vnd.google-apps.spreadsheet",
		Parents: []string{newFolder.Id},
	})
	errorPrinter(err)
	if createdFile != nil {
		prettyPrinter(fmt.Sprintf("Created File ID: %s", createdFile.Id))
	}
	
	filePath := getGoDotEnvVariable("FILE_PATH")
	file, err := os.Open(filePath)
	errorPrinter(err)
	uploadedFile, err := uploadFiles(file, parentFolderUrl)
	errorPrinter(err)
	file.Close()
	prettyPrinter( fmt.Sprintf("File Uploaded: %s", uploadedFile.Name) )
	
	files, err := getFolderInfos(parentFolderUrl)
	errorPrinter(err)
	for index, file := range files {
		fmt.Printf(`
		[%d] %s (%s)
		-----------`, index, file.Name, file.Id)

		if strings.Contains(strings.ToLower(file.Name), "grade") {
			copiedFile, err := copyFileTo(file, newFolder.Id)
			errorPrinter(err)
			prettyPrinter(fmt.Sprintf("This is the copied file:\n%#v", copiedFile.Id))

			targetFolderUrl := getGoDotEnvVariable("OTHER_FOLDER_URL")

			movedFile, err := moveFileTo(parentFolderUrl, targetFolderUrl, file)
			errorPrinter(err)
			prettyPrinter(fmt.Sprintf("This is the moved file:\n%s", movedFile.Id))

			downloadFiles(file, "C:\\dev", "pdf")
		}

		if strings.Contains(strings.ToLower(file.Name), "captura") {
			permanentlyDeleteFile(file.Id)
		}
	}

	emptyTrash()
}