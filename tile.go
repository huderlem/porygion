package porygion

// Tile is a 8x8-pixel section in a region map.
type Tile struct {
	X, Y int
}

// Distance computes the manhattan distance between two Tiles.
func (t Tile) Distance(other Tile) int {
	xDiff := t.X - other.X
	if xDiff < 0 {
		xDiff = -xDiff
	}
	yDiff := t.Y - other.Y
	if yDiff < 0 {
		yDiff = -yDiff
	}
	return xDiff + yDiff
}
