package models

type EventType string

// Event types that the backend can report
const (
	EventCreated EventType = "created"
	EventUpdated EventType = "updated"
	EventDeleted EventType = "deleted"
)

type EventResource string

const (
	EventResourceUser       EventResource = "user"
	EventResourceConnection EventResource = "connection"
)

// Event is used to send notifications to k8s layer for reconciliation.
// As we are not embedding updated resource itself into the event,
// Component, consuming the event, should fetch the updated resource from the database.
type Event struct {
	ID       string        `json:"-" yaml:"-" gorm:"primaryKey"`
	Type     EventType     `json:"type"`
	Resource EventResource `json:"resource"`
	ObjectID string        `json:"objectId"`
}
