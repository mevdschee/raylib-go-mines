package movies

import (
	"encoding/json"
	"fmt"

	"github.com/mevdschee/raylib-go-mines/clips"
	"github.com/mevdschee/raylib-go-mines/scenes"
	"github.com/mevdschee/raylib-go-mines/sprites"
)

// Movie is a set of scenes
type Movie struct {
	currentScene *scenes.Scene
	scenes       map[string]*scenes.Scene
}

// New creates a new movie
func New() *Movie {
	return &Movie{
		currentScene: nil,
		scenes:       map[string]*scenes.Scene{},
	}
}

// FromJSON creates a new movie from JSON
func FromJSON(spriteMap sprites.SpriteMap, data string, parameters map[string]interface{}) (*Movie, error) {
	sceneJSONs := []scenes.SceneJSON{}
	err := json.Unmarshal([]byte(data), &sceneJSONs)
	if err != nil {
		return nil, err
	}
	movie := Movie{
		currentScene: &scenes.Scene{},
		scenes:       map[string]*scenes.Scene{},
	}
	for _, sceneJSON := range sceneJSONs {
		scene, err := scenes.FromJSON(spriteMap, sceneJSON, parameters)
		if err != nil {
			return nil, err
		}
		movie.Add(scene)
	}
	return &movie, nil
}

// Add adds a scene to the movie
func (m *Movie) Add(scene *scenes.Scene) {
	m.scenes[scene.GetName()] = scene
	if len(m.scenes) == 1 {
		m.currentScene = scene
	}
}

// Draw draws the movie
func (m *Movie) Draw(scale int) {
	if m.currentScene != nil {
		m.currentScene.Draw(scale)
	}
}

// Update updates the movie
func (m *Movie) Update(scale int) (err error) {
	if m.currentScene != nil {
		err = m.currentScene.Update(scale)
	}
	return err
}

// GetClip gets a clip from the movie
func (m *Movie) GetClip(scene, layer, clip string) (*clips.Clip, error) {
	return m.getClip(scene, layer, clip, 0)
}

// GetClip gets a clip from the movie
func (m *Movie) getClip(scene, layer, clip string, i int) (*clips.Clip, error) {
	if s, ok := m.scenes[scene]; ok {
		return s.GetClip(layer, clip, i)
	}
	return nil, fmt.Errorf("GetClip: scene '%s' not found", scene)
}

// GetClips gets a series of clips from the movie
func (m *Movie) GetClips(scene, layer, clip string) ([]*clips.Clip, error) {
	clips := []*clips.Clip{}
	for i := 0; true; i++ {
		c, err := m.getClip(scene, layer, clip, i)
		if err != nil {
			if i == 0 {
				return clips, err
			}
			break
		}
		clips = append(clips, c)
	}
	return clips, nil
}
