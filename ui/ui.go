package ui

type GoclipLauncher interface {
	ShowEntries()
	RedrawApps()
}

type GoclipSettings interface {
	SetReloadAppsCallback(callback func())
	ShowSettings()
	Run()
}
