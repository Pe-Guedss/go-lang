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

// Gets the service used to make every drive operation
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

// Checks if an error is null.
// If it is not, this function prints it.
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

// Prints a message in a more distinguible way.
func prettyPrinter (msgs... string) {
	for _, msg := range(msgs) {
		fmt.Printf(`
	-------------------
	%s
	-------------------`, msg)
	}
}

// Retrieves environment variables from the ".env" file stored in this repo.
func getGoDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")
  
	if err != nil {
	  log.Fatalf("Error loading .env file")
	}
  
	return os.Getenv(key)
}

// Checks for file duplicates inside a folder. If it finds one, the return will be true.
// 
// Please note that the file search is based in name and type, so independently of dates and other metadata, if two files have the same name and type, it will return true.
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

// Searches inside a folder for a file duplicate, when it finds, return the file found.
// If no file is found, the return is a nil pointer.
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

// Retrieves the ID from a drive URL. Firstly, it checks for the "https" prefix, if it does not have one, the function just returns the url given, assuming that it is already an ID.
func getFolderId (url string) string {
	if strings.HasPrefix(url, "https"){
		arr := strings.Split(url, "folders/")

		return arr[1]
	}

	return url
}

// Returns all the files inside a folder. You must provide an drive folder url or ID.
func getFolderFiles (url string) (*drive.FileList, error){
	folderId := getFolderId(url)
	query := fmt.Sprintf("parents = '%s'", folderId)

	fileList, err := srv.Files.List().Q(query).Do()

	return fileList, err
}

// Retrives an array of files from the "drive.FileList" structure. It checks for any additional pages and retrieves the files from those as well.
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

// This function creates a new folder inside a parent. You have to provide the name of the folder to be created, along with the parent URL or ID.
// 
// Please note that this function checks for duplicates. So, if there already is a folder inside the parent with the same name, it will not create a new folder.
// 
// This function also returns the folder created, so, if there is a duplicate, it will return the folder that already exists.
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

// This function copies a file inside a parent. You have to provide the file to be copied, as well as the destination folder ID.
// 
// Please note that this function checks for duplicates. So, if there already is a file inside the parent with the same name and time, it will not make the copy.
// 
// This function also returns the file copied, so, if there is a duplicate, it will return the file that already exists.
func copyFileTo (file *drive.File, destinationFolderId string) (*drive.File, error) {

	isDuplicate, err := checkFileDuplicates(file, destinationFolderId)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		prettyPrinter("This file already exists inside the parent folder.")
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

// This function creates a new file inside a parent. You have to provide the file that will be created. Remember to add the parent ID to the respective file struct field.
// 
// Please note that this function checks for duplicates. So, if there already is a file inside the parent with the same name, it will not create a new file.
// 
// This function also returns the file created, so, if there is a duplicate, it will return the file that already exists.
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

// This function moves a file to a given parent. You have to provide the source of the folder, it's destination folder, as well as the file you want to move.
// 
// Please note that this function *DOES NOT* checks for duplicates. So, if there already is a file inside the parent with the same name, it will move the file anyways.
func moveFileTo (source string, target string, file *drive.File) (*drive.File, error) {
	sourceId := getFolderId(source)
	targetId := getFolderId(target)

	movedFile, err := srv.Files.Update(
		file.Id,
		&drive.File{},
	).AddParents(targetId).RemoveParents(sourceId).Do()

	return movedFile, err
}

// This function uploads a local file to a given drive parent. You have to provide the local file as an "os.File" pointer and the destination drive folder that you want to upload the file.
// 
// To get the local file, you can use: file, err := os.Open(filePath). Do not forget to close the file afterwards.
// 
// Please note that this function *DOES NOT* checks for duplicates. So, if there already is a file inside the parent with the same name, it will upload the new file anyways.
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

// This function downloads a drive file to a given local parent. You have to provide the drive file, the downloaded file format and the destination local folder that you want to upload the file.
// 
// Please note that this function *DOES NOT* checks for duplicates in the local folder. So, if there already is a file inside the folder with the same name, it will download the new file anyways.
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

// This functions permanently deletes a file. You must provide the file ID to do so.
// 
// Please note that this function *DOES NOT* moves the file to the trash, it just deletes it and you cannot retrieve it anymore.
func permanentlyDeleteFile (fileId string) error {
	err := srv.Files.Delete(fileId).Do()
	return err
}

// This function permanently deletes all the files inside the trash.
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