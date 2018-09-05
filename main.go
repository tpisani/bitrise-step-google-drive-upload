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
	artifactPath := os.Getenv("artifact_path")
	if artifactPath == "" {
		fmt.Fprintln(os.Stderr, "artifact_path not defined!")
		os.Exit(1)
	}

	artifactName := os.Getenv("artifact_name")
	if artifactName == "" {
		parts := strings.Split(artifactPath, "/")
		artifactName = parts[len(parts)-1]
	}

	clientID := os.Getenv("google_drive_client_id")
	if clientID == "" {
		fmt.Fprintln(os.Stderr, "google_drive_client_id is not defined!")
		os.Exit(1)
	}

	clientSecret := os.Getenv("google_drive_client_secret")
	if clientSecret == "" {
		fmt.Fprintln(os.Stderr, "google_drive_client_secret is not defined!")
		os.Exit(1)
	}

	refreshToken := os.Getenv("google_drive_refresh_token")
	if refreshToken == "" {
		fmt.Fprintln(os.Stderr, "google_drive_refresh_token is not defined!")
		os.Exit(1)
	}

	folderID := os.Getenv("google_drive_folder_id")
	if folderID == "" {
		fmt.Fprintln(os.Stderr, "google_drive_folder_id is not defined!")
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
		fmt.Fprintf(os.Stderr, "Unable to list folder (%s): %v\n", folderID, err)
		os.Exit(1)
	}

	var fileID *string
	for _, fp := range fileList.Files {
		if fp.Name == artifactName {
			fileID = &fp.Id
			break
		}
	}

	fp := &drive.File{Name: artifactName}

	if fileID == nil {
		fp.Parents = []string{folderID}
		_, err = svc.Files.Create(fp).Media(f).Do()
	} else {
		_, err = svc.Files.Update(*fileID, fp).AddParents(folderID).Media(f).Do()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to upload file: %v\n", err)
		os.Exit(1)
	}
}
