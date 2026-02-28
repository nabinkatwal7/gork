package main

import (
	"image"
	"os"
	"path/filepath"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font/opentype"
)

// assetPaths returns paths to try for a given asset file (e.g. "ui_bg.png").
// Tries assets/ and ./assets/ so the game works from project root or other dirs.
func assetPaths(name string) []string {
	return []string{
		filepath.Join("assets", name),
		filepath.Join(".", "assets", name),
	}
}

type Assets struct {
	Background      *ebiten.Image
	AppBar          *ebiten.Image
	Panel           *ebiten.Image
	PanelInset      *ebiten.Image
	ButtonPrimary   *ebiten.Image
	ButtonSecondary *ebiten.Image
	ButtonGhost     *ebiten.Image
	Chip            *ebiten.Image
	ListRow         *ebiten.Image
	SceneFrame      *ebiten.Image
	MinimapBG       *ebiten.Image
	WorldMapBG      *ebiten.Image
	Icons           *ebiten.Image
	Font            *opentype.Font
}

func LoadAssets() *Assets {
	assets := &Assets{}
	assets.Background = loadImageFirst(assetPaths("ui_bg.png"))
	assets.AppBar = loadImageFirst(assetPaths("appbar_bg.png"))
	assets.Panel = loadImageFirst(assetPaths("panel_9slice.png"))
	assets.PanelInset = loadImageFirst(assetPaths("panel_inset_9slice.png"))
	assets.ButtonPrimary = loadImageFirst(assetPaths("button_primary_9slice.png"))
	assets.ButtonSecondary = loadImageFirst(assetPaths("button_secondary_9slice.png"))
	assets.ButtonGhost = loadImageFirst(assetPaths("button_ghost_9slice.png"))
	assets.Chip = loadImageFirst(assetPaths("chip_9slice.png"))
	assets.ListRow = loadImageFirst(assetPaths("listrow_9slice.png"))
	assets.SceneFrame = loadImageFirst(assetPaths("scene_frame.png"))
	assets.MinimapBG = loadImageFirst(assetPaths("minimap_bg.png"))
	assets.WorldMapBG = loadImageFirst(assetPaths("worldmap_bg.png"))
	assets.Icons = loadImageFirst(assetPaths("icons.png"))
	assets.Font = loadFontFirst(assetPaths("font.ttf"))
	return assets
}

func loadImageFirst(paths []string) *ebiten.Image {
	for _, path := range paths {
		if img := loadImage(path); img != nil {
			return img
		}
	}
	return nil
}

func loadFontFirst(paths []string) *opentype.Font {
	for _, path := range paths {
		if f := loadFont(path); f != nil {
			return f
		}
	}
	return nil
}

func loadImage(path string) *ebiten.Image {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(img)
}

func loadFont(path string) *opentype.Font {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	font, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	return font
}
