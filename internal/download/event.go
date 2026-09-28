package download

// DownloadEvent models real-time progress and lifecycle events broadcasted to listeners.
type DownloadEvent struct {
	TaskID     string  `json:"task_id"`
	Title      string  `json:"title"`
	Filename   string  `json:"filename"`
	State      State   `json:"state"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Progress   float64 `json:"progress"`
	Speed      int64   `json:"speed"` // Bytes per second
	Error      error   `json:"-"`
}

// Listener is a callback invoked when a download state or progress update occurs.
type Listener func(event DownloadEvent)
