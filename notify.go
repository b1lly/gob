package gob

import "github.com/mattn/go-gntp"

var (
	growlClient  *gntp.Client
	inErrorState = false
)

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
	if inErrorState {
		// TODO(ttacon): add dynamic body text, icon and callback
		growlClient.Notify(&gntp.Message{
			Event: "fixed",
			Title: "Build Fixed",
			Text:  "The build is fixed!",
		})
		inErrorState = false
	}
}

func notifyFailed() {
	if !inErrorState {
		// TODO(ttacon): same todo as above
		growlClient.Notify(&gntp.Message{
			Event: "failed",
			Title: "Build Failed",
			Text:  "The build has failed",
		})
		inErrorState = true
	}
}
