package main

import (
	"bytes"
	"encoding/binary"
	. "github.com/stojg/vivere/lib/components"
	"time"
	"strings"
)

var (
	entities       *EntityManager
	modelList      *ModelList
	collisionList  *CollisionList
	rigidList      *RigidBodyList
	controllerList *ControllerList
)

func NewLevel(monitor *Monitor) *Level {
	entities = NewEntityManager()
	modelList = NewModelList()
	rigidList = NewRigidBodyManager()
	collisionList = NewCollisionList()
	controllerList = NewControllerList()

	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for {
			Println("Updating instances")
			monitor.UpdateInstances()
			Println("Instances updated")
			<-ticker.C
		}
	}()

	lvl := &Level{}
	lvl.systems = append(lvl.systems, &PhysicSystem{})
	lvl.systems = append(lvl.systems, &ControllerSystem{})
	lvl.systems = append(lvl.systems, &CollisionSystem{})
	return lvl
}

type Level struct {
	systems []System
}

func (l *Level) Update(elapsed float64) {
	for i := range l.systems {
		l.systems[i].Update(elapsed)
	}
}

func (l *Level) Draw() *bytes.Buffer {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, float32(Frame))

	for id, component := range modelList.All() {
		binaryStream(buf, INST_ENTITY_ID, *id)
		binaryStream(buf, INST_SET_POSITION, component.Position)
		binaryStream(buf, INST_SET_ORIENTATION, component.Orientation)
		binaryStream(buf, INST_SET_TYPE, component.Model)
		binaryStream(buf, INST_SET_SCALE, component.Scale)
		inst := monitor.FindByEntityID(*id)
		var health float64
		if inst.CPUCreditBalance < 10 && strings.HasPrefix(inst.InstanceType, "t2") && inst.State == "running" {
			health = 0.0
		} else {
			health = 1.0 - inst.CPUUtilization/100.0
		}
		binaryStream(buf, INST_SET_HEALTH, health)
		//Printf("cpu %f  health %f", inst.CPUUtilization, health)
	}

	return buf
}
