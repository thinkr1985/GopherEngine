package widgets

import rl "github.com/gen2brain/raylib-go/raylib"

var widget_default_font rl.Font

func initializeWidgetFont() {
	widget_default_font = rl.LoadFontEx("fonts/CALIBRI.TTF", 12, nil, 0)
	// defer rl.UnloadFont(widget_default_font)
}
