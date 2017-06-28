package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"google.golang.org/api/sheets/v4"

	"html/template"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/mux"
)

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := reader.ReadString('\n')
	fmt.Println(text)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), err
}
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
func getClient1(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) *Page {
	filename := title + ".txt"
	body, _ := ioutil.ReadFile(filename)
	return &Page{Title: title, Body: body}
}

func subWrap(srv *sheets.Service, id string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()
		parameters := r.URL.Query()
		//Inizialmente salvo i dati inseriti dalle textbox (data,operatore,ecc)
		rb := &sheets.ValueRange{
			Values: [][]interface{}{{parameters["Data"][0]}, {parameters["Operatore"][0]}, {parameters["nome_scheda"][0]}, {parameters["numero_unita"][0]}},
		}
		srv.Spreadsheets.Values.Update(id, "B2:B", rb).ValueInputOption("RAW").Context(ctx).Do()

		for k, v := range parameters {
			if strings.Contains(k, "radio") { //Ogni radio button ha un suffisso che indica la relativa pos nel doc
				//Questo suffisso viene parsato per ottenere il range in notazione A1
				vRange := strings.TrimPrefix(k, "radio")
				rb := &sheets.ValueRange{
					Values: [][]interface{}{{v[0]}},
				}
				srv.Spreadsheets.Values.Update(id, vRange+":"+vRange, rb).ValueInputOption("RAW").Context(ctx).Do()
			}
		}
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p := loadPage(title)
	p = &Page{Title: title}
	t, _ := template.ParseFiles("template/visual_inspection.html")
	t.Execute(w, p)
}

func main() {

	//-----AUTHENTICATION---BEGIN
	ctx := context.Background()
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient1(ctx, config)
	srv, err := sheets.New(client)

	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	//AUTHENTICATION---END
	spreadSheetID := "1tDJDInHsC8IlP6CgxJNctAwZjn9yKpnQTFxP9LOXSbI"
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/submit", subWrap(srv, spreadSheetID))
	var port string
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Print(port)
	http.ListenAndServe(":"+port, r)

}
