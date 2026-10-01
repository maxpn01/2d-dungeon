package rendering

type Visibility interface{ EntityVisible(x, y, size float64) bool }
