package ui

type GoclipLauncher interface {
	ShowEntries()
	Start()
}

type GoclipSettings interface {
	ShowSettings()
	Start()
}
