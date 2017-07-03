package inventory

import (
	"time"
)

type Base struct {
	Time 	time.Time
	Subject	interface{} 		// Subject of what is going on (e.g. container)
	Action	string
	Object  interface{}         // Passive object
	Tags 	map[string]string 	// Tags that should be applied to the action
}
