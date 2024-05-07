package sprites

import (
	"bytes"
	"encoding/json"
	"image/png"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SpriteMap is a map of sprites
type SpriteMap map[string]*Sprite

// Sprite is the base struct for any sprite
type Sprite struct {
	Image   *rl.Image    `json:"-"`
	Texture rl.Texture2D `json:"-"`
	Name    string       `json:"name"`
	X       int          `json:"x"`
	Y       int          `json:"y"`
	Width   int          `json:"width,omitempty"`
	Height  int          `json:"height,omitempty"`
	Widths  [3]int       `json:"widths,omitempty"`
	Heights [3]int       `json:"heights,omitempty"`
	Count   int          `json:"count"`
	Grid    int          `json:"grid"`
	Gap     int          `json:"gap,omitempty"`
}

// NewSpriteMap creates a new sprite map
func NewSpriteMap(imagedata []byte, jsondata string) (SpriteMap, error) {
	image, err := png.Decode(bytes.NewReader(imagedata))
	if err != nil {
		return nil, err
	}
	sprites := []*Sprite{}
	spriteMap := SpriteMap{}
	err = json.Unmarshal([]byte(jsondata), &sprites)
	if err != nil {
		return nil, err
	}
	spriteImage := rl.NewImageFromImage(image)
	spriteTexture := rl.LoadTextureFromImage(spriteImage)
	for _, sprite := range sprites {
		sprite.Image = spriteImage
		sprite.Texture = spriteTexture
		spriteMap[sprite.Name] = sprite
	}
	return spriteMap, nil
}
