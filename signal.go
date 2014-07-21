package gob

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// this function allows us to tell gob to restart the process it is running
func registerSignalHandlers(g *Gob) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		for _ = range c {
			// waiting for CTRL-C
			select {
			case <-time.After(time.Second):
				// TODO(ttacon): ensure we kill child process
				g.Print("\r[gob] exiting...")
				os.Exit(0)
			case <-c:
				select {
				case <-time.After(time.Millisecond * 300):
					g.restartApp()
				case <-c:
					// TODO(ttacon): ensure we kill child process
					g.Print("\r[gob] exiting...")
					os.Exit(0)
				}
			}
		}
	}()
}
