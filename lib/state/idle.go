package state

import (
	. "github.com/stojg/cyberspace/lib/components"
	. "github.com/stojg/steering"
	"github.com/stojg/vector"
	. "github.com/stojg/vivere/lib/components"
)

func NewIdle(e *Entity, i *AWSInstance, m *Model, r *RigidBody) *Idle {
	return &Idle{
		entity:   e,
		instance: i,
		model:    m,
		body:     r,
	}
}

type Idle struct {
	entity   *Entity
	instance *AWSInstance
	body     *RigidBody
	model    *Model
}

func (s *Idle) Update() State {
	positions := siblings(s.instance, s.model, true)

	if len(positions) < 2 {
		return nil
	}

	midpoint := &vector.Vector3{}
	for i := range positions {
		midpoint.Add(positions[i])
	}
	midpoint.Scale(1 / float64(len(positions)))

	if midpoint.NewSub(s.model.Position()).SquareLength() > (s.model.Scale[0]*s.model.Scale[0])*2 {
		//fmt.Printf(" %s switching to cluster (%0.2f)\n", s.instance.Name, midpoint.NewSub(s.model.Position()).Length())
		return NewCluster(s.entity, s.instance, s.model, s.body, midpoint)
	}

	return nil
}

func (s *Idle) Steering() *SteeringOutput {
	if s.instance.CPUUtilization() < 5.0 {
		return nil
	}
	if s.body.Rotation.Length() < s.instance.CPUUtilization()/80 {
		steer := NewSteeringOutput()
		steer.SetAngular(&vector.Vector3{0, 1, 0})
		return steer
	}
	return nil
}
