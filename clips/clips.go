package clips

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/mevdschee/raylib-go-mines/sprites"
)

// Clip is a set of frames
type Clip struct {
	name             string
	x, y             float32
	width, height    float32
	frame            int
	frames           []rl.Texture2D
	onPress          func()
	onLongPress      func()
	onRelease        func()
	onReleaseOutside func()
}

// ClipJSON is a clip in JSON
type ClipJSON struct {
	Name          string
	Sprite        string
	Repeat        string
	X, Y          string
	Width, Height string
}

// GetName gets the name of the clip
func (c *Clip) GetName() string {
	return c.name
}

// New creates a new sprite based clip
func New(sprite *sprites.Sprite, name string, x, y float32) *Clip {
	frames := []rl.Texture2D{}

	srcWidth, srcHeight := sprite.Width, sprite.Height
	for i := 0; i < sprite.Count; i++ {
		grid := sprite.Grid
		if grid == 0 {
			grid = sprite.Count
		}
		srcX := sprite.X + (i%grid)*(srcWidth+sprite.Gap)
		srcY := sprite.Y + (i/grid)*(srcHeight+sprite.Gap)
		r := rl.NewRectangle(float32(srcX), float32(srcY), float32(srcWidth), float32(srcHeight))
		frame := rl.NewImageFromImage(sprite.Image)
		rl.ImageCrop(frame, r)
		frames = append(frames, rl.LoadTextureFromImage(frame))
		rl.UnloadImage(frame)
	}

	return &Clip{
		name:   name,
		x:      x,
		y:      y,
		width:  float32(srcWidth),
		height: float32(srcHeight),
		frame:  0,
		frames: frames,
	}
}

// NewScaled creates a new 9 slice scaled sprite based clip
func NewScaled(sprite *sprites.Sprite, name string, x, y, width, height int) *Clip {
	frame0 := rl.NewImageFromImage(image.NewNRGBA(image.Rect(0, 0, width, height)))
	srcImage := rl.NewImageFromImage(sprite.Image)

	srcY := sprite.Y
	dstY := 0
	for h := 0; h < 3; h++ {
		srcHeight := sprite.Heights[h]
		dstHeight := sprite.Heights[h]
		if h == 1 {
			dstHeight = height - sprite.Heights[0] - sprite.Heights[2]
		}
		srcX := sprite.X
		dstX := 0
		for w := 0; w < 3; w++ {
			srcWidth := sprite.Widths[w]
			dstWidth := sprite.Widths[w]
			if w == 1 {
				dstWidth = width - sprite.Widths[0] - sprite.Widths[2]
			}

			srcRect := rl.NewRectangle(float32(srcX), float32(srcY), float32(srcWidth), float32(srcHeight))
			dstRect := rl.NewRectangle(float32(dstX), float32(dstY), float32(dstWidth), float32(dstHeight))
			rl.ImageDraw(frame0, srcImage, srcRect, dstRect, rl.White)
			srcX += srcWidth + sprite.Gap
			dstX += dstWidth
		}
		srcY += srcHeight + sprite.Gap
		dstY += dstHeight
	}

	frames := []rl.Texture2D{rl.LoadTextureFromImage(frame0)}
	rl.UnloadImage(frame0)
	rl.UnloadImage(srcImage)

	return &Clip{
		name:   name,
		x:      float32(x),
		y:      float32(y),
		width:  float32(width),
		height: float32(height),
		frame:  0,
		frames: frames,
	}
}

// Draw draws the clip
func (c *Clip) Draw() {
	img := c.frames[c.frame]
	rl.DrawTextureEx(img, rl.NewVector2(c.x, c.y), 0, 1, rl.White)
}

// GotoFrame goes to a frame of the clip
func (c *Clip) GotoFrame(frame int) {
	if frame >= 0 && frame < len(c.frames) {
		c.frame = frame
	}
}

// OnPress sets the click handler function
func (c *Clip) OnPress(handler func()) {
	c.onPress = handler
}

// OnLongPress sets the click handler function
func (c *Clip) OnLongPress(handler func()) {
	c.onLongPress = handler
}

// OnRelease sets the click handler function
func (c *Clip) OnRelease(handler func()) {
	c.onRelease = handler
}

// OnReleaseOutside sets the click handler function
func (c *Clip) OnReleaseOutside(handler func()) {
	c.onReleaseOutside = handler
}

// IsHovered returns whether or not the cursor is hovering the clip
func (c *Clip) IsHovered() bool {
	cursor := rl.GetMousePosition()
	rect := rl.NewRectangle(c.x, c.y, c.width, c.height)
	return rl.CheckCollisionPointRec(cursor, rect)
}

// // IsTouched returns whether or not the touch hits the clip
// func (c *Clip) IsTouched(touchID ebiten.TouchID) bool {
// 	cursorX, cursorY := ebiten.TouchPosition(touchID)
// 	cursor := image.Point{cursorX, cursorY}
// 	rect := image.Rect(c.x, c.y, c.x+c.width, c.y+c.height)
// 	return cursor.In(rect)
// }

// // IsTouched returns whether or not the touch hits the clip
// func (c *Clip) IsTouchedPreviously(touchID ebiten.TouchID) bool {
// 	cursorX, cursorY := inpututil.TouchPositionInPreviousTick(touchID)
// 	cursor := image.Point{cursorX, cursorY}
// 	rect := image.Rect(c.x, c.y, c.x+c.width, c.y+c.height)
// 	return cursor.In(rect)
// }

// Update updates the clip
func (c *Clip) Update() (err error) {
	hover := c.IsHovered()

	if c.onPress != nil {
		if hover && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			c.onPress()
		}
	}
	if c.onLongPress != nil {
		if hover && rl.IsMouseButtonPressed(rl.MouseRightButton) {
			c.onLongPress()
		}
	}
	if c.onRelease != nil {
		if hover && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			c.onRelease()
		}
	}
	if c.onReleaseOutside != nil {
		if !hover && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			c.onReleaseOutside()
		}
	}
	// touchIDs := touch.GetTouchIDs()
	// for i := 0; i < len(touchIDs); i++ {
	// 	touchID := touchIDs[i]
	// 	touched := c.IsTouched(touchID)
	// 	touchedPreviously := c.IsTouchedPreviously(touchID)
	// 	if c.onPress != nil {
	// 		if touched && touch.IsTouchJustPressed(touchID) {
	// 			c.onPress()
	// 		}
	// 	}
	// 	if c.onLongPress != nil {
	// 		if touched && inpututil.TouchPressDuration(touchID) == ebiten.TPS()/2 {
	// 			c.onLongPress()
	// 		}
	// 	}
	// 	if c.onRelease != nil {
	// 		if touchedPreviously && touch.IsTouchJustReleased(touchID) {
	// 			c.onRelease()
	// 		}
	// 	}
	// 	if c.onReleaseOutside != nil {
	// 		if !touchedPreviously && inpututil.IsTouchJustReleased(touchID) {
	// 			c.onReleaseOutside()
	// 		}
	// 	}
	//}
	return nil
}
