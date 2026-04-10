package realtime

import (
	"sync"
	"time"
)

const (
	EventTypeTaskCreated = "task_created"
	EventTypeTaskUpdated = "task_updated"
	EventTypeTaskDeleted = "task_deleted"
)

const subscriberBufferSize = 16

type TaskPayload struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ProjectID   string    `json:"project_id"`
	AssigneeID  *string   `json:"assignee_id"`
	CreatorID   string    `json:"creator_id"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Event struct {
	Type      string       `json:"type"`
	ProjectID string       `json:"project_id"`
	Task      *TaskPayload `json:"task,omitempty"`
	TaskID    *string      `json:"task_id,omitempty"`
}

type Publisher interface {
	Publish(event Event)
}

// Manager is a small in-memory pub/sub broker for project-scoped task events.
// Slow subscribers are dropped so task writes never block on streaming clients.
type Manager struct {
	mu          sync.RWMutex
	nextID      uint64
	subscribers map[string]map[uint64]chan Event
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[string]map[uint64]chan Event),
	}
}

func (m *Manager) Subscribe(projectID string) (<-chan Event, func()) {
	ch := make(chan Event, subscriberBufferSize)

	m.mu.Lock()
	m.nextID++
	subscriberID := m.nextID
	if _, exists := m.subscribers[projectID]; !exists {
		m.subscribers[projectID] = make(map[uint64]chan Event)
	}
	m.subscribers[projectID][subscriberID] = ch
	m.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			m.removeSubscriber(projectID, subscriberID)
		})
	}

	return ch, unsubscribe
}

func (m *Manager) Publish(event Event) {
	m.mu.RLock()
	projectSubscribers, exists := m.subscribers[event.ProjectID]
	if !exists || len(projectSubscribers) == 0 {
		m.mu.RUnlock()
		return
	}

	blockedSubscriberIDs := make([]uint64, 0)
	for subscriberID, ch := range projectSubscribers {
		select {
		case ch <- event:
		default:
			blockedSubscriberIDs = append(blockedSubscriberIDs, subscriberID)
		}
	}
	m.mu.RUnlock()

	if len(blockedSubscriberIDs) == 0 {
		return
	}

	for _, subscriberID := range blockedSubscriberIDs {
		m.removeSubscriber(event.ProjectID, subscriberID)
	}
}

func (m *Manager) removeSubscriber(projectID string, subscriberID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectSubscribers, exists := m.subscribers[projectID]
	if !exists {
		return
	}

	ch, exists := projectSubscribers[subscriberID]
	if !exists {
		return
	}

	delete(projectSubscribers, subscriberID)
	close(ch)

	if len(projectSubscribers) == 0 {
		delete(m.subscribers, projectID)
	}
}
