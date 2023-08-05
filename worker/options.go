package worker

import "github.com/alitto/pond"

// resizerFromString returns a pond resizer for the given name.
func resizerFromString(name string) pond.ResizingStrategy {
	switch name {
	case "eager":
		return pond.Eager()
	case "lazy":
		return pond.Lazy()
	case "balanced":
		return pond.Balanced()
	default:
		panic("invalid resizer name")
	}
}
