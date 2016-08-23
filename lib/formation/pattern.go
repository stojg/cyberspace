package formation

type Pattern interface {
	SupportsSlots(int) bool
	SlotLocation(int) Static
	DriftOffset([]*SlotAssignment) Static
}
