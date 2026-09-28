package settings

// Repository decouples settings persistence from UI controls and business logic.
type Repository interface {
	Load() (Settings, error)
	Save(Settings) error
	Reset() error
}
