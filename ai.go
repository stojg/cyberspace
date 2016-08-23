package main

import (
	. "github.com/stojg/cyberspace/lib/components"
	. "github.com/stojg/cyberspace/states"
	. "github.com/stojg/vivere/lib/components"
	"github.com/stojg/vivere/lib/vector"
)

func NewAI(ent *Entity, model *Model, body *RigidBody) *AI {
	inst := monitor.FindByEntityID(*ent)
	ai := &AI{
		instance: inst,
		entity:   ent,
		model:    model,
		body:     body,
		state:    NewIdle(ent, inst, model, body),
	}
	return ai
}

type AI struct {
	entity   *Entity
	instance *Instance
	model    *Model
	body     *RigidBody
	target   *vector.Vector3
	state    State
	reminder float64
}

func (s *AI) Update(elapsed float64) {

	//we only update the state every 100ms
	s.reminder += elapsed
	if s.reminder > 0.1 {
		s.reminder = -0.1
		state := s.state.Update()
		if state != nil {
			s.state = state
		}
	}

	if s.state == nil {
		s.state = NewIdle(s.entity, s.instance, s.model, s.body)
	}

	steering := s.state.Steering()
	if steering != nil {
		s.body.AddForce(steering.Linear())
		s.body.AddTorque(steering.Angular())
	}

	// clamp to the ground
	s.model.Position()[1] = s.model.Scale[1] / 2
}
