// Takes care of OS X notification center pushing
// Author: Chun Li

package gob

import (
	"log"

	"github.com/deckarep/gosx-notifier"
)

// push all notifications through a central channel
// this way if we need any pre or post push hooks, we can easily add them
var Notify = make(chan *gosxnotifier.Notification)

func ListenForNotifications() {
	for note := range Notify {
		PrePush(note)
		err := note.Push()
		if err != nil {
			log.Printf("Error pushing the notification %v: %v", note, err)
		}
	}
}

// PrePush is run before a notification gets pushed out.
// It contains some default parts of our message we might always want.
func PrePush(note *gosxnotifier.Notification) {
	if note.Title != "" {
		note.Title = "Gob"
	}
	note.Group = "gob"
	note.AppIcon = "gob_icon.png" // Note: requires Mavericks (OS X 10.9+)
}

// Say pushes out a notification with the specified message.
func Say(message string) {
	Notify <- &gosxnotifier.Notification{Message: message}
}
