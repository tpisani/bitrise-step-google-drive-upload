package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func main() {
	artifactPath := os.Getenv("ARTIFACT_PATH")
	if artifactPath == "" {
		fmt.Fprintln(os.Stderr, "$ARTIFACT_PATH not defined!")
		os.Exit(1)
	}

	artifactName := os.Getenv("ARTIFACT_NAME")
	if artifactName == "" {
		parts := strings.Split(artifactPath, "/")
		artifactName = parts[len(parts)-1]
	}

	clientID := os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
	if clientID == "" {
		fmt.Fprintln(os.Stderr, "$GOOGLE_DRIVE_CLIENT_ID is not defined!")
		os.Exit(1)
	}

	clientSecret := os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET")
	if clientSecret == "" {
		fmt.Fprintln(os.Stderr, "$GOOGLE_DRIVE_CLIENT_SECRET is not defined!")
		os.Exit(1)
	}

	refreshToken := os.Getenv("GOOGLE_DRIVE_REFRESH_TOKEN")
	if refreshToken == "" {
		fmt.Fprintln(os.Stderr, "$GOOGLE_DRIVE_REFRESH_TOKEN is not defined!")
		os.Exit(1)
	}

	folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID")
	if folderID == "" {
		fmt.Fprintln(os.Stderr, "$GOOGLE_DRIVE_FOLDER_ID is not defined!")
		os.Exit(1)
	}

	f, err := os.Open(artifactPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open artifact: %v\n", err)
		os.Exit(1)
	}

	t := &oauth2.Token{RefreshToken: refreshToken}
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			drive.DriveScope,
		},
	}
	client := config.Client(context.Background(), t)

	svc, err := drive.New(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize Google Drive service: %v\n", err)
		os.Exit(1)
	}

	fileList, err := svc.Files.List().Q(fmt.Sprintf("'%s' in parents", folderID)).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to list folder: %v\n", err)
		os.Exit(1)
	}

	var fileID *string
	for _, fp := range fileList.Files {
		if fp.Name == artifactName {
			fileID = &fp.Id
			break
		}
	}

	if fileID == nil {
		fp := &drive.File{Name: artifactName, Parents: []string{folderID}}
		_, err = svc.Files.Create(fp).Media(f).Do()
	} else {
		fp := &drive.File{Name: artifactName}
		_, err = svc.Files.Update(*fileID, fp).AddParents(folderID).Media(f).Do()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to upload file: %v\n", err)
		os.Exit(1)
	}
}
