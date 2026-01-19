package config

type ManagerMock struct {
    Cfg Config
}

func (m *ManagerMock) Config() Config    { return m.Cfg }
func (m *ManagerMock) Hot() HotConfig    { return extractHot(m.Cfg) }
func (m *ManagerMock) Path() string      { return "" }
func (m *ManagerMock) ReloadIfNeeded(onError func(error)) {}
func (m *ManagerMock) EnsureLogDir() error { return nil }
func (m *ManagerMock) ResolvePath(p string) string { return p }
