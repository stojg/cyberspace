package formation

import (
	. "github.com/stojg/vivere/lib/vector"
)

type Static interface {
	Position() *Vector3
	Orientation() *Quaternion
}

type Model struct {
	position    *Vector3
	orientation *Quaternion
}

func (m *Model) Position() *Vector3 {
	return m.position
}

func (m *Model) Orientation() *Quaternion {
	return m.orientation
}
