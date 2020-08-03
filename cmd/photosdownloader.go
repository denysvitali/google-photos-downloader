package main

import (
	"context"
	"fmt"
	"github.com/alexflint/go-arg"
	gphotos "github.com/denysvitali/google-photos-api-client-go/v2"
	"github.com/denysvitali/photos-downloader/pkg/handlers"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"
)

var args struct {
	ClientId     string `arg:"env:CLIENT_ID,--client-id" help:"API's Client ID'"`
	ClientSecret string `arg:"env:CLIENT_SECRET,--client-secret" help:"API's Client Secret'"`
	AlbumId      string `arg:"positional,required" help:"Album ID"`
	OutputPath   string `arg:"-o,--output,required"`
}

const (
	HttpPort = 5837
)

type ConfigFile struct {
	Token *oauth2.Token `yaml:"token"`
}

func main() {
	arg.MustParse(&args)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create output dir

	err := os.MkdirAll(args.OutputPath, 0744)

	if err != nil {
		logrus.Fatal(err)
	}

	currentUser, err := user.Current()
	if err != nil {
		logrus.Fatal(err)
	}

	err = os.MkdirAll(path.Join(currentUser.HomeDir, ".config", "photosdownloader"), 0744)
	if err != nil {
		logrus.Fatal(err)
	}

	file, err := os.OpenFile(
		path.Join(currentUser.HomeDir, ".config", "photosdownloader", "config.yml"),
		os.O_CREATE|os.O_RDWR, 0600)

	if err != nil {
		logrus.Fatal(err)
	}

	configFile := ConfigFile{}
	err = yaml.NewDecoder(file).Decode(&configFile)

	var token *oauth2.Token

	if err != nil {
		// Create Oauth Session
		token = getOauthToken(args.ClientId, args.ClientSecret)

		// Write token to file

		configFile.Token = token
		err = yaml.NewEncoder(file).Encode(&configFile)
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		token = configFile.Token
	}

	uploaderSessionStore := gphotos.MemoryUploadSessionStore{}
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client, err := gphotos.NewClientWithOptions(tc, uploaderSessionStore, gphotos.WithLogger(logger))
	if err != nil {
		logrus.Fatal(err)
	}

	album, err := client.AlbumById(ctx, args.AlbumId)

	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("album: %v\n", album)

	mediaItems, err := client.MediaItemsByAlbum(ctx, album, 1000)
	if err != nil {
		logrus.Fatal(err)
	}

	var wg sync.WaitGroup

	for _, v := range mediaItems {
		if strings.HasPrefix(v.MimeType, "image/") {
			// Download image
			downloadUrl := v.BaseUrl + "=d"
			wg.Add(1)
			go downloadPicture(downloadUrl, args.OutputPath, v.FileName, &wg, logger)
		} else {
			logger.Debugf("Invalid mime: %s", v.MimeType)
		}
	}

	wg.Wait()
}

func downloadPicture(url string, outputPath string, name string, wg *sync.WaitGroup, logger *logrus.Logger) {
	logger.Debugf("downloading %s", name)
	file, err := os.OpenFile(path.Join(outputPath, name), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Errorf("unable to create file: %v", err)
		wg.Done()
		return
	}

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		logger.Errorf("unable to get URL: %v", err)
		wg.Done()
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Errorf("unable to copy src to dest: %v", err)
		wg.Done()
		return
	}
	wg.Done()
}

func getOauthToken(clientId string, clientSecret string) *oauth2.Token {
	handler := handlers.New(clientId, clientSecret, HttpPort)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", HttpPort),
		Handler: handler.GetHandler(),
	}

	go httpServer.ListenAndServe()

	// Wait for authentication
	fmt.Printf("Server up and running, please visit http://127.0.0.1:%d", HttpPort)

	for {
		if handler.HasToken() {
			break
		}
		time.Sleep(1 * time.Second)
	}
	token := handler.GetToken()
	httpServer.Close()
	return token
}
