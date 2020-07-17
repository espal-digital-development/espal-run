package system

// System provides tools to communicate with the core/OS system.
type System struct {
}

// SetSoftUlimit sets the OS's native soft max and cur limit.
func (s *System) SetSoftUlimit(max uint64, cur uint64) error {
	return s.setSoftUlimit(max, cur)
}

// New returns a new instance of System.
func New() (*System, error) {
	s := &System{}
	return s, nil
}
