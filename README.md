# Raylib Go Mines

Implementation of minesweeper in Go using the Raylib game library.

![menu](screenshot_menu.png)

![game](screenshot_game.png)

### Requirements

On Debian based systems: 

    sudo apt install build-essential libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev mingw-w64

### Running

To run the software execute:

    go mod tidy
    go run .

First build may take several minutes.

### Building

In order to install the resource bundler run:

    go install github.com/tc-hib/go-winres@latest

To build the software with bundled resources execute:

    make

First build may take several minutes.