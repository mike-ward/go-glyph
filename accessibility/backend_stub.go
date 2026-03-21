package accessibility

// stubBackend is the no-op fallback for all platforms.
// Platform-specific files (backend_darwin.go, backend_linux.go)
// override newBackend and newAnnouncerBackend via build tags.
type stubBackend struct{}

func (stubBackend) UpdateTree(map[int]Node, int)            {}
func (stubBackend) SetFocus(int)                            {}
func (stubBackend) PostNotification(int, Notification)      {}
func (stubBackend) UpdateTextField(int, string, Range, int) {}
func (stubBackend) Flush()                                  {}

type stubAnnouncerBackend struct{}

func (stubAnnouncerBackend) Announce(string) {}

// newBackend returns the platform accessibility backend.
// Overridden by build-tagged files on darwin/linux.
func newBackend() Backend {
	return stubBackend{}
}

// newAnnouncerBackend returns the platform announcer backend.
func newAnnouncerBackend() AnnouncerBackend {
	return stubAnnouncerBackend{}
}
