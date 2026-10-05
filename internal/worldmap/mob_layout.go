package worldmap

import (
	"math/rand/v2"
	"strings"
)

func randomizeSideRoomGuards(layout *Layout, seed uint64) {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	rooms := []string{WestArchive, WestGallery, EastArmory, EastGallery}
	pairs := [][2]MobType{
		{DungeonPursuer, DungeonPursuer},
		{DungeonHeavy, DungeonHeavy},
		{DungeonShooter, DungeonHeavy},
		{DungeonShooter, DungeonPursuer},
	}
	rng.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	assignments := make(map[string][2]MobType, len(rooms))
	for i, id := range rooms {
		pair := pairs[i]
		if rng.IntN(2) == 0 {
			pair[0], pair[1] = pair[1], pair[0]
		}
		assignments[id] = pair
	}
	for i := range layout.Rooms {
		room := &layout.Rooms[i]
		pair, ok := assignments[strings.TrimPrefix(room.ID, "north-")]
		if !ok {
			continue
		}
		slot := 0
		for j := range room.FeedingAreas {
			guard := room.FeedingAreas[j].Guard
			if guard != nil {
				guard.Type = pair[slot%len(pair)]
				slot++
			}
		}
	}
}
