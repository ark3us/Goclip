package ui

type GoclipLauncher interface {
	ShowEntries()
	RedrawApps()
}

type GoclipSettings interface {
	ShowSettings()
	Run()
}
