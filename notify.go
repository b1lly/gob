package gob

import "github.com/mattn/go-gntp"

var growlClient *gntp.Client

func initNotifications() {
	growlClient = gntp.NewClient()
	growlClient.AppName = "gob"
	growlClient.Register([]gntp.Notification{
		gntp.Notification{
			Event:   "fixed",
			Enabled: false,
		}, gntp.Notification{
			Event:   "failed",
			Enabled: true,
		},
	})
}

func notifyFixed() {
	// TODO(ttacon): add dynamic body text, icon and callback
	growlClient.Notify(&gntp.Message{
		Event: "fixed",
		Title: "Build Fixed",
		Text:  "The build is fixed!",
	})
}

func notifyFailed() {
	// TODO(ttacon): same todo as above
	growlClient.Notify(&gntp.Message{
		Event: "failed",
		Title: "Build Failed",
		Text:  "The build has failed",
	})
}
