package widgets

import rl "github.com/gen2brain/raylib-go/raylib"

var Widget_default_font rl.Font
var Widget_body_font rl.Font

func InitializeWidgetFont() {
	Widget_default_font = rl.LoadFontEx("fonts/CALIBRIB.TTF", 14, nil, 0)
	Widget_body_font = rl.LoadFontEx("fonts/CALIBRI.TTF", 14, nil, 0)
	// defer rl.UnloadFont(widget_default_font)
}
