package rendering

func drawPlayerSplitCells(p Actor, exploration Visibility) {
	count := p.Split.Count()
	for _, cell := range p.Split {
		if cell.Active {
			if exploration == nil || exploration.EntityVisible(cell.X, cell.Y, p.Size) {
				DrawSplitBody(cell.X, cell.Y, p.Size, cell.Stats.HP, p.Tint, p.Character, count)
			}
		}
	}
}

func drawAISplitCells(ai Actor, exploration Visibility) {
	count := ai.Split.Count()
	for _, cell := range ai.Split {
		if cell.Active && (exploration == nil || exploration.EntityVisible(cell.X, cell.Y, ai.Size)) {
			DrawSplitBody(cell.X, cell.Y, ai.Size, cell.Stats.HP, ai.Tint, ai.Character, count)
		}
	}
}
