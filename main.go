package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"log"
	"math/rand"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mevdschee/raylib-go-mines/clips"
	"github.com/mevdschee/raylib-go-mines/movies"
	"github.com/mevdschee/raylib-go-mines/sprites"
)

//go:embed winxpskin.png
var spriteMapImage []byte

//go:embed minesicon.png
var minesIconImage []byte

const spriteMapMeta = `
	[{"name":"display","x":28,"y":82,"width":41,"height":25,"count":1},
	{"name":"icons","x":0,"y":0,"width":16,"height":16,"count":17,"grid":9},
	{"name":"digits","x":0,"y":33,"width":11,"height":21,"count":11,"gap":1},
	{"name":"buttons","x":0,"y":55,"width":26,"height":26,"count":5,"gap":1},
	{"name":"controls","x":0,"y":82,"widths":[12,1,12],"heights":[11,1,11],"gap":1},
	{"name":"field","x":0,"y":96,"widths":[12,1,12],"heights":[11,1,11],"gap":1}]`

const movieScenes = `
	[{"name":"game","layers":[{"name":"bg","clips":[
		{"sprite":"controls","x":"0","y":"0","width":"w*16+24","height":"55"},
		{"sprite":"field","x":"0","y":"44","width":"w*16+24","height":"h*16+22"},
		{"sprite":"display","x":"16","y":"15"},
		{"sprite":"display","x":"w*16-33","y":"15"}
	]},{"name":"fg","clips":[
		{"sprite":"digits","name":"bombs","repeat":"3","x":"18+i*13","y":"17"},
		{"sprite":"digits","name":"time","repeat":"3","x":"w*16-31+i*13","y":"17"},
		{"sprite":"buttons","name":"button","x":"(w*16)/2-1","y":"15"},
		{"sprite":"icons","name":"icons","repeat":"w*h","x":"12+(i%w)*16","y":"55+floor(i/w)*16"}
	]}]}]`

type config struct {
	scale   int
	width   int
	height  int
	bombs   int
	holding int
}

type game struct {
	c      config
	movie  *movies.Movie
	button int
	bombs  int
	closed int
	state  int
	time   int64
	tiles  [][]tile
}

type tile struct {
	open    bool
	marked  bool
	bomb    bool
	pressed bool
	number  int
}

const (
	stateWaiting = iota
	statePlaying
	stateWon
	stateLost
)

const (
	buttonPlaying = iota
	buttonEvaluate
	buttonLost
	buttonWon
	buttonPressed
)

const (
	iconEmpty = iota
	iconNumberOne
	iconNumberTwo
	iconNumberThree
	iconNumberFour
	iconNumberFive
	iconNumberSix
	iconNumberSeven
	iconNumberEight
	iconClosed
	iconOpened
	iconBomb
	iconMarked
	iconAnswerNoBomb
	iconAnswerIsBomb
	iconQuestionMark
	iconQuestionPressed
)

var clipCache map[string][]*clips.Clip

var version string

func (g *game) getSize() (int, int) {
	return g.c.width*16 + 12*2, g.c.height*16 + 11*3 + 33
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.getSize()
}

func (g *game) init() {
	spriteMap, err := sprites.NewSpriteMap(spriteMapImage, spriteMapMeta)
	if err != nil {
		log.Fatalln(err)
	}
	parameters := map[string]interface{}{
		"w": g.c.width,
		"h": g.c.height,
	}
	movie, err := movies.FromJSON(spriteMap, movieScenes, parameters)
	if err != nil {
		log.Fatalln(err)
	}
	g.movie = movie
	clipCache = map[string][]*clips.Clip{}
}

func (g *game) getClips(clip string) []*clips.Clip {
	if clipCache == nil {
		clipCache = map[string][]*clips.Clip{}
	}
	cache, ok := clipCache[clip]
	if ok {
		return cache
	}
	clips, err := g.movie.GetClips("game", "fg", clip)
	if err != nil {
		log.Fatal(err)
	}
	clipCache[clip] = clips
	return clips
}

func (g *game) setHandlers() {
	button := g.getClips("button")[0]
	button.OnPress(func() {
		g.button = buttonPressed
	})
	button.OnRelease(func() {
		if g.button == buttonPressed {
			g.restart()
		}
	})
	button.OnReleaseOutside(func() {
		if g.button == buttonPressed {
			g.restart()
		}
	})
	icons := g.getClips("icons")
	for y := 0; y < g.c.height; y++ {
		for x := 0; x < g.c.width; x++ {
			px, py := x, y
			icons[y*g.c.width+x].OnPress(func() {
				if g.state == stateWon || g.state == stateLost {
					return
				}
				if g.tiles[py][px].marked {
					return
				}
				g.button = buttonEvaluate
				g.tiles[py][px].pressed = true
				if g.tiles[py][px].open {
					g.forEachNeighbour(px, py, func(x, y int) {
						if !g.tiles[y][x].marked {
							g.tiles[y][x].pressed = true
						}
					})
				}
			})
			icons[y*g.c.width+x].OnLongPress(func() {
				if g.state == stateWon || g.state == stateLost {
					return
				}
				g.onPressTile(px, py, true)
				g.tiles[py][px].pressed = false
			})
			icons[y*g.c.width+x].OnRelease(func() {
				if g.state == stateWon || g.state == stateLost {
					return
				}
				g.button = buttonPlaying
				if g.tiles[py][px].open {
					g.onPressTile(px, py, true)
				} else {
					if g.tiles[py][px].pressed {
						g.onPressTile(px, py, false)
					}
				}
				g.tiles[py][px].pressed = false
			})
			icons[y*g.c.width+x].OnReleaseOutside(func() {
				if g.state == stateWon || g.state == stateLost {
					return
				}
				g.button = buttonPlaying
				g.tiles[py][px].pressed = false
			})
		}
	}
}

func (g *game) forEachNeighbour(x, y int, do func(x, y int)) {
	for i := 0; i < 9; i++ {
		dy, dx := i/3-1, i%3-1
		if dy == 0 && dx == 0 {
			continue
		}
		if y+dy < 0 || x+dx < 0 {
			continue
		}
		if y+dy >= g.c.height || x+dx >= g.c.width {
			continue
		}
		do(x+dx, y+dy)
	}
}

func (g *game) onPressTile(x, y int, long bool) {
	if g.state == stateWaiting {
		g.state = statePlaying
		g.time = time.Now().UnixNano()
		g.placeBombs(x, y, g.bombs)
	}
	if !long && g.tiles[y][x].marked {
		return
	}
	if g.tiles[y][x].open {
		if long {
			var marked = 0
			g.forEachNeighbour(x, y, func(x, y int) {
				if g.tiles[y][x].marked {
					marked++
				}
			})
			if g.tiles[y][x].number == marked {
				g.forEachNeighbour(x, y, func(x, y int) {
					if !g.tiles[y][x].marked {
						g.onPressTile(x, y, false)
					}
				})
			}
		}
	} else {
		if long {
			if g.tiles[y][x].marked {
				g.tiles[y][x].marked = false
				g.bombs++
			} else {
				g.tiles[y][x].marked = true
				g.bombs--
			}
		} else {
			g.tiles[y][x].open = true
			g.closed--
			if g.tiles[y][x].bomb {
				g.state = stateLost
				g.button = buttonLost
				return
			}
			if g.tiles[y][x].number == 0 {
				g.forEachNeighbour(x, y, func(x, y int) {
					g.onPressTile(x, y, false)
				})
			}
		}
	}
}

func (g *game) setButton() {
	button := g.getClips("button")[0]
	button.GotoFrame(g.button)
}

func (g *game) setNumbers() {
	bombsDigits := g.getClips("bombs")
	bombs := g.bombs
	if g.state == stateWon {
		bombs = 0
	}
	if bombs < -99 {
		bombs = -99
	}
	negative := false
	if bombs < 0 {
		negative = true
		bombs *= -1
	}
	for i := 0; i < 3; i++ {
		if i == 2 && negative {
			bombsDigits[2-i].GotoFrame(10)
		} else {
			bombsDigits[2-i].GotoFrame(bombs % 10)
		}
		bombs /= 10
	}
	if g.state == statePlaying || g.state == stateWaiting {
		time := int((time.Now().UnixNano() - g.time) / 1000000000)
		if time > 999 {
			time = 999
		}
		timeDigits := g.getClips("time")
		for i := 0; i < 3; i++ {
			timeDigits[2-i].GotoFrame(time % 10)
			time /= 10
		}
	}
}

func (g *game) setTiles() {
	icons := g.getClips("icons")
	if g.state == stateWon || g.state == stateLost {
		for y := 0; y < g.c.height; y++ {
			for x := 0; x < g.c.width; x++ {
				icon := iconClosed
				if g.tiles[y][x].open {
					if g.tiles[y][x].bomb {
						icon = iconAnswerIsBomb
					} else {
						icon = g.tiles[y][x].number
					}
				} else {
					if g.tiles[y][x].marked {
						if g.tiles[y][x].bomb {
							icon = iconMarked
						} else {
							icon = iconAnswerNoBomb
						}
					} else {
						if g.tiles[y][x].bomb {
							if g.state == stateWon {
								icon = iconMarked
							} else {
								icon = iconBomb
							}
						}
					}
				}
				icons[y*g.c.width+x].GotoFrame(icon)
			}
		}
	} else {
		for y := 0; y < g.c.height; y++ {
			for x := 0; x < g.c.width; x++ {
				icon := iconClosed
				if g.tiles[y][x].open {
					icon = g.tiles[y][x].number
				} else {
					if g.tiles[y][x].marked {
						icon = iconMarked
					} else {
						if g.tiles[y][x].pressed {
							icon = iconEmpty
						}
						if g.state == stateWon {
							icon = iconMarked
						}
					}
				}
				icons[y*g.c.width+x].GotoFrame(icon)
			}
		}
	}
}

func (g *game) Update(scale int) error {
	if g.movie == nil {
		g.init()
		g.setHandlers()
	}
	if g.state == stateWaiting {
		g.time = time.Now().UnixNano()
	}
	g.setButton()
	g.setNumbers()
	g.setTiles()
	if g.state == statePlaying {
		if g.closed == g.c.bombs {
			g.state = stateWon
			g.button = buttonWon
		}
	}
	//touch.UpdateTouchIDs()
	return g.movie.Update(scale)
}

func (g *game) Draw(scale int) {
	g.movie.Draw(scale)
}

func newGame(c config) *game {
	g := &game{c: c}
	return g
}

func (g *game) restart() {
	g.button = buttonPlaying
	g.bombs = g.c.bombs
	g.closed = g.c.width * g.c.height
	g.state = stateWaiting
	g.time = time.Now().UnixNano()
	g.tiles = make([][]tile, g.c.height)
	for y := 0; y < g.c.height; y++ {
		g.tiles[y] = make([]tile, g.c.width)
		for x := 0; x < g.c.width; x++ {
			g.tiles[y][x] = tile{}
		}
	}
}

func (g *game) placeBombs(x, y, bombs int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := bombs
	g.tiles[y][x].bomb = true
	for b > 0 {
		x, y := rng.Intn(g.c.width), rng.Intn(g.c.height)
		if !g.tiles[y][x].bomb {
			g.tiles[y][x].bomb = true
			b--
			g.forEachNeighbour(x, y, func(x, y int) {
				g.tiles[y][x].number++
			})
		}
	}
	g.tiles[y][x].bomb = false
}

func main() {
	//rl.SetTraceLog(rl.LogError)
	title := "Raylib Go Mines v" + version
	c := config{
		scale:   2,
		width:   9,
		height:  9,
		bombs:   10,
		holding: 15,
	}
	g := newGame(c)
	menu := true
	g.restart()
	width, height := g.getSize()
	rl.InitWindow(int32(c.scale*width), int32(c.scale*height), title)
	rl.SetTargetFPS(30)
	icon, err := png.Decode(bytes.NewReader(minesIconImage))
	if err == nil {
		rl.SetWindowIcon(*rl.NewImageFromImage(icon))
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.White)
		if menu {
			gui.SetStyleProperty(gui.GlobalTextFontsize, int64(g.c.scale*10))
			w := float32(g.c.scale * width)
			h := float32(g.c.scale * height)
			m := float32(g.c.scale * 5)
			cy := m
			row := float32(g.c.scale * 20)
			gui.Label(rl.NewRectangle(m, cy, w-2*m, row), title)
			cy += row + m
			beginner := gui.Button(rl.NewRectangle(m, cy, w-2*m, row), "Beginner")
			if beginner {
				c.height = 9
				c.width = 9
				c.bombs = 10
			}
			cy += row + m/2
			intermediate := gui.Button(rl.NewRectangle(m, cy, w-2*m, row), "Intermediate")
			if intermediate {
				c.height = 16
				c.width = 16
				c.bombs = 40
			}
			cy += row + m/2
			expert := gui.Button(rl.NewRectangle(m, cy, w-2*m, row), "Expert")
			if expert {
				c.height = 16
				c.width = 30
				c.bombs = 99
			}
			cy += row + m + m
			//gui.Label(rl.NewRectangle(m, cy, 0, row), "Scale:")
			//c.scale = gui.Spinner(rl.NewRectangle(w/2-m, cy, w/2-m, row), c.scale, 1, 6)
			//cy += row + m
			gui.Label(rl.NewRectangle(m, cy, 0, row), "Height:")
			c.height = gui.Spinner(rl.NewRectangle(w/2-m, cy, w/2-m, row), c.height, 9, 50)
			cy += row + m
			gui.Label(rl.NewRectangle(m, cy, 0, row), "Width:")
			c.width = gui.Spinner(rl.NewRectangle(w/2-m, cy, w/2-m, row), c.width, 9, 100)
			cy += row + m
			gui.Label(rl.NewRectangle(m, cy, 0, row), "Bombs:")
			c.bombs = gui.Spinner(rl.NewRectangle(w/2-m, cy, w/2-m, row), c.bombs, 1, 999)
			start := gui.Button(rl.NewRectangle(m, h-m-row, w-2*m, row), "Start")
			if start {
				g.c = c
				g.restart()
				width, height := g.getSize()
				rl.SetWindowSize(g.c.scale*width, g.c.scale*height)
				rl.SetWindowPosition((rl.GetMonitorWidth(0)-g.c.scale*width)/2, (rl.GetMonitorHeight(0)-g.c.scale*height)/2)
				menu = false
			}
		} else {
			g.Update(g.c.scale)
			g.Draw(g.c.scale)
		}
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
