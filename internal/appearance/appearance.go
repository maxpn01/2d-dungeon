package appearance

type Character string

const (
	Square   Character = "square"
	Triangle Character = "triangle"
)

func Enabled(c Character) bool {
	return c == Square
}

func Normalize(c Character) Character {
	if Enabled(c) {
		return c
	}
	return Square
}
