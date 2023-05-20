package diagrams

const (
	DirectionDown Direction = iota
	DirectionRight
)

type Direction int64

func (d Direction) String() string {
	switch d {
	case DirectionRight:
		return "right"
	default:
		return "down"
	}
}
