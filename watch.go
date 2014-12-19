package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-fsnotify/fsnotify"
)

func watch(c config) error {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}
	defer watcher.Close()

	const dir = "./"

	if err := watcher.Add(dir); err != nil {
		return err
	}

	interval := 1 * time.Second
	tick := time.Tick(interval)
	done := make(chan struct{})

	// a buffer for events
	var events []fsnotify.Event
	var loopErr error

	go func() {
	Loop:
		for {
			select {
			case event := <-watcher.Events:
				fmt.Println("watcher.Events")
				loopErr = errors.New("shenanigans")
				break Loop

				if !strings.HasSuffix(event.Name, ".go") {
					continue
				}
				if is(event, fsnotify.Create) || is(event, fsnotify.Write) {
					events = append(events, event)
				}
			case loopErr = <-watcher.Errors:
				fmt.Println("watcher.Errors")
				done <- struct{}{}
			case <-tick:
				fmt.Println("tick")
				if len(events) == 0 {
					continue
				}

				// stop watching while gen'ing files
				loopErr = watcher.Remove(dir)
				if loopErr != nil {
					break Loop
				}

				// gen the files
				loopErr = run(c)
				if loopErr != nil {
					break Loop
				}

				// clear the buffer
				events = make([]fsnotify.Event, 0)

				// resume watching
				loopErr = watcher.Add(dir)
				if loopErr != nil {
					break Loop
				}
			}
		}
		done <- struct{}{}
	}()

	s := <-done
	fmt.Println(s)
	close(done)
	fmt.Println("closed")

	if loopErr != nil {
		fmt.Println("returning an err")
		fmt.Println(loopErr)
		return loopErr
	}

	fmt.Println("returning nil")
	return nil
}

func is(event fsnotify.Event, op fsnotify.Op) bool {
	return event.Op&op == op
}
