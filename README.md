# Goclip

Simple Windows-like clipboard manager for linux.

This application is just a proof-of-concept and might be highly unstable.

## Features

- Clipboard text and image support (https://github.com/golang-design/clipboard)
- App launcher
- Shell commands with autocomplete
- Data persistence (https://github.com/asdine/storm)
- System shortcut (https://github.com/robotn/gohook)
- Simple UI (https://github.com/gotk3/gotk3)
- System tray icon (https://github.com/getlantern/systray)


## Usage

AppImage executable is available in Releases.

### Default hotkeys

- Win+V : open clipboard manager
- Win+C : open app launcher
- Win+x : open shell launcher

### Clipbord manager shortcuts

- Left click: copy entry into clipboard
- Right click: open entry with default app

### App launcher shortcuts

- Left click: launch application

### Shell launcher shortcuts

- Focus suggestion: autocomplete
- Enter in search box: execute in default terminal
- Right click on entry: execute in terminal

## Build

### Requirements
```
# For https://github.com/robotn/gohook 
sudo apt install gcc libc6-dev
sudo apt install libx11-dev xorg-dev libxtst-dev libpng++-dev
sudo apt install xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev
sudo apt install xsel xclip
# For https://github.com/golang-design/clipboard
sudo apt install libx11-dev
# For https://github.com/gotk3/gotk3
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev
# For https://github.com/getlantern/systray
sudo apt-get install gcc libgtk-3-dev libappindicator3-dev
```

## Improvements

A LOT can be improved, this is just a proof-of-concept...

- Memory management
- Interface
- ...
