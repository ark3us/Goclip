# Goclip

Simple Windows-like clipboard manager for linux.

This application is just a proof-of-concept and might be highly unstable.

## Features

- Clipboard text and image support (https://github.com/golang-design/clipboard)
- App launcher
- Data persistence (https://github.com/asdine/storm)
- System shortcut (https://github.com/robotn/gohook)
- Simple UI (https://github.com/gotk3/gotk3)
- System tray icon (https://github.com/getlantern/systray)

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

# Usage

## Default keys

- Alt+z : open clipboard manager
- Alt+x : open app launcher

# Improvements

A LOT can be improved, this is just a proof-of-concept...

- Memory management
- Interface
- ...