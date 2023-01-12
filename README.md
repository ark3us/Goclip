# Goclip

Simple Windows-like clipboard manager and command launcher for Linux / Ubuntu, written in Go.

This application is just a proof-of-concept and might be highly unstable.

## Features

- Clipboard manager 
- App launcher
- Shell launcher with autocomplete
- Text and image support (https://github.com/golang-design/clipboard)
- Data persistence (https://github.com/asdine/storm)
- System shortcut (https://github.com/robotn/gohook)
- Gtk3 UI (https://github.com/gotk3/gotk3)
- System tray (https://github.com/fyne-io/systray)


## Usage

### Default hotkeys

- Alt+V : open clipboard manager
- Alt+C : open app launcher
- Alt+x : open shell launcher

### Clipbord manager shortcuts

- Left click: copy entry into clipboard
- Right click: open entry with default app

### App launcher shortcuts

- Left click: launch application

### Shell launcher shortcuts

- Focus suggestion: autocomplete
- Enter in search box: execute in default terminal
- Right click on entry: execute in terminal

#### Supported terminals:

- gnome-terminal
- terminator

## Build

Pre-built AppImage executable available here: https://github.com/ark3us/Goclip/releases

### Requirements
```
# For https://github.com/robotn/gohook 
sudo apt install xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev
# For https://github.com/golang-design/clipboard
sudo apt install libx11-dev
# For https://github.com/gotk3/gotk3
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev
```

## Improvements

A LOT can be improved, this is just a proof-of-concept...

- Memory management
- Interface
- ...
