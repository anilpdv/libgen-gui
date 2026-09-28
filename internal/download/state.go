package download

import "fmt"

// State defines the discrete lifecycle states for a download task.
type State uint8

const (
	StateQueued State = iota
	StateDownloading
	StatePaused
	StateCompleted
	StateFailed
	StateCancelled
)

func (s State) String() string {
	switch s {
	case StateQueued:
		return "queued"
	case StateDownloading:
		return "downloading"
	case StatePaused:
		return "paused"
	case StateCompleted:
		return "completed"
	case StateFailed:
		return "failed"
	case StateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// UserLabel returns a formatted, capitalized label for UI presentation.
func (s State) UserLabel() string {
	switch s {
	case StateQueued:
		return "Queued"
	case StateDownloading:
		return "Downloading"
	case StatePaused:
		return "Paused"
	case StateCompleted:
		return "Completed"
	case StateFailed:
		return "Failed"
	case StateCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// Legal state transition table.
var transitions = map[State]map[State]bool{
	StateQueued: {
		StateDownloading: true,
		StatePaused:      true,
		StateCancelled:   true,
	},
	StateDownloading: {
		StatePaused:    true,
		StateCompleted: true,
		StateFailed:    true,
		StateCancelled: true,
	},
	StatePaused: {
		StateQueued:      true,
		StateDownloading: true,
		StateCancelled:   true,
	},
	StateFailed: {
		StateQueued:      true,
		StateDownloading: true,
		StateCancelled:   true,
	},
	StateCancelled: {
		StateQueued: true, // Allow retrying cancelled tasks
	},
	StateCompleted: {
		// Terminal state: completed tasks cannot be resumed without re-enqueuing
	},
}

// ValidateTransition returns an error if transitioning from current to next is illegal.
func ValidateTransition(current, next State) error {
	if validStates, ok := transitions[current]; ok && validStates[next] {
		return nil
	}
	return fmt.Errorf("invalid state transition: %s -> %s", current, next)
}
