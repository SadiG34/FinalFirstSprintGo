package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Event struct {
	Type    string
	Message string
	Time    time.Time
}

type EventLogger struct {
	ch chan Event
}

func NewEventLogger() *EventLogger {
	l := &EventLogger{
		ch: make(chan Event, 100),
	}
	go l.worker()
	return l
}

func (l *EventLogger) Log(eventType, message string) {
	l.ch <- Event{Type: eventType, Message: message, Time: time.Now()}
}

func (l *EventLogger) Close() {
	close(l.ch)
}

func (l *EventLogger) worker() {
	file, err := os.OpenFile("events.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("ERROR: Failed to open log file: %v", err)
		for event := range l.ch {
			fmt.Fprintf(os.Stderr, "[LOGGER ERROR] %s: %s\n", event.Type, event.Message)
		}
		return
	}
	defer file.Close()

	writer := log.New(file, "", log.LstdFlags)
	for event := range l.ch {
		line := fmt.Sprintf("[%s] %s: %s", event.Time.Format("15:04:05"), event.Type, event.Message)
		fmt.Println(line)
		writer.Println(line)
	}
}