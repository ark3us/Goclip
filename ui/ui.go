package ui

type GoclipLauncher interface {
	ShowEntries()
}

type GoclipSettings interface {
	ShowSettings()
	Run()
}
