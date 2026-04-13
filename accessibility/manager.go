package accessibility

// Backend is the platform interface for the accessibility tree.
type Backend interface {
	UpdateTree(nodes map[int]Node, rootID int)
	SetFocus(nodeID int)
	PostNotification(nodeID int, n Notification)
	UpdateTextField(nodeID int, value string, sel Range, cursorLine int)
	Flush()
}

// Manager manages the accessibility tree lifecycle.
type Manager struct {
	backend Backend
	nodes   map[int]Node
	nextID  int
	rootID  int
}

// NewManager creates a manager with platform-specific backend.
func NewManager() *Manager {
	return &Manager{
		backend: newBackend(),
		nodes:   make(map[int]Node),
		nextID:  1,
	}
}

// AddTextNode adds a text node under the root.
func (m *Manager) AddTextNode(text string, rect Rect) {
	if len(m.nodes) == 0 {
		m.reset()
	}
	id := m.nextNodeID()
	m.nodes[id] = Node{
		ID:     id,
		Role:   RoleText,
		Rect:   rect,
		Text:   text,
		Parent: m.rootID,
	}
	if root, ok := m.nodes[m.rootID]; ok {
		root.Children = append(root.Children, id)
		m.nodes[m.rootID] = root
	}
}

// CreateTextFieldNode creates an editable text field node.
func (m *Manager) CreateTextFieldNode(rect Rect) int {
	if len(m.nodes) == 0 {
		m.reset()
	}
	id := m.nextNodeID()
	m.nodes[id] = Node{
		ID:     id,
		Role:   RoleTextField,
		Rect:   rect,
		Parent: m.rootID,
	}
	if root, ok := m.nodes[m.rootID]; ok {
		root.Children = append(root.Children, id)
		m.nodes[m.rootID] = root
	}
	return id
}

// UpdateTextField updates text field attributes via backend.
func (m *Manager) UpdateTextField(nodeID int, value string,
	sel Range, cursorLine int) {
	m.backend.UpdateTextField(nodeID, value, sel, cursorLine)
}

// SetFocus notifies the backend of focus change.
func (m *Manager) SetFocus(nodeID int) {
	m.backend.SetFocus(nodeID)
}

// PostNotification posts an accessibility notification.
func (m *Manager) PostNotification(nodeID int, n Notification) {
	m.backend.PostNotification(nodeID, n)
}

// Flush processes pending platform events.
func (m *Manager) Flush() {
	m.backend.Flush()
}

// Commit pushes accumulated updates then resets.
func (m *Manager) Commit() {
	if len(m.nodes) == 0 {
		return
	}
	m.backend.UpdateTree(m.nodes, m.rootID)
	m.reset()
}

func (m *Manager) reset() {
	clear(m.nodes)
	m.nextID = 1
	m.rootID = m.nextNodeID()
	m.nodes[m.rootID] = Node{
		ID:   m.rootID,
		Role: RoleContainer,
		Text: "Content",
	}
}

func (m *Manager) nextNodeID() int {
	id := m.nextID
	m.nextID++
	return id
}
