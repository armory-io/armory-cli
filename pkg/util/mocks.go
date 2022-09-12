package util

import (
	"errors"
)

type Mock struct {
	Calls      int
	CalledWith string
}

func NewMock() Mock {
	return Mock{Calls: 0, CalledWith: ""}
}

func (m *Mock) Inc() {
	m.Calls = m.Calls + 1
}
func (m *Mock) Errors() error {
	m.Calls = m.Calls + 1
	return errors.New("some error")
}

func (m *Mock) ErrorsWith(s string) error {
	m.CalledWith = m.CalledWith + ";" + s
	return errors.New("some error: " + s)
}

func (m *Mock) Updates() error {
	m.Calls = m.Calls + 1
	return nil
}

func (m *Mock) UpdatesWith(s string) error {
	m.CalledWith = m.CalledWith + ";" + s
	return nil
}
