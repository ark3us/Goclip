package ui

type GoclipLauncher interface {
	ShowEntries()
	Run()
}

type GoclipSettings interface {
	ShowSettings()
	Run()
}
