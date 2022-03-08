package ui

type GoclipLauncher interface {
	ShowEntries()
	RedrawApps()
	RedrawClipboardHistory()
}

type GoclipSettings interface {
	SetReloadAppsCallback(callback func())
	ShowSettings()
	Run()
}
