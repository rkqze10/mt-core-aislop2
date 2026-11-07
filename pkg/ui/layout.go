package ui

func (m *AppModel) calculateLayout() (boxWidth, vpHeight int, isCompact bool) {
	w := m.width
	if w <= 0 {
		w = 80
	}
	h := m.height
	if h <= 0 {
		h = 24
	}

	boxWidth = w - 4
	if boxWidth > 140 {
		boxWidth = 140
	}
	if boxWidth < 20 {
		boxWidth = w - 2
		if boxWidth < 10 {
			boxWidth = 10
		}
	}

	isCompact = h < 32 || w < 70 || (m.hasWorkspace && len(m.tabs) > 0)

	overhead := 14
	if isCompact {
		overhead = 10
	}
	if h < 22 {
		overhead = 7
	}
	if h < 14 {
		overhead = 4
	}

	vpHeight = h - overhead
	if vpHeight < 2 {
		vpHeight = 2
	}
	if vpHeight > 60 {
		vpHeight = 60
	}
	return boxWidth, vpHeight, isCompact
}
