package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitializeFirebase() *firebase.App {
	// You need to download your Firebase service account key JSON file and place it in the config directory
	opt := option.WithCredentialsFile("config/firebase-service-account.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase: %v\n", err)
		return nil
	}

	return app
}
