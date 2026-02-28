package ime

// stubBridge is the no-op fallback for all platforms.
// Platform-specific files override newPlatformBridge via build tags.
type stubBridge struct{}

func (stubBridge) Enable(float32, float32, float32, float32) {}
func (stubBridge) Disable()                                   {}
func (stubBridge) SetCursorRect(float32, float32, float32, float32) {}
func (stubBridge) IsActive() bool                              { return false }

func newPlatformBridge(_ Callbacks) Bridge {
	return stubBridge{}
}
