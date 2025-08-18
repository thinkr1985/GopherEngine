package guis

import (
	"GopherEngine/core"
	"GopherEngine/widgets"
)

type SceneExplorerPanel struct {
	TreeViewer *widgets.TreeWidget
	Scene      *core.Scene
}

func NewSceneExplorer(scene *core.Scene) *SceneExplorerPanel {
	return &SceneExplorerPanel{
		TreeViewer: widgets.NewTreeWidget(10, 10, 200, 100, 12),
		Scene:      scene,
	}

}

func (se *SceneExplorerPanel) Update() {

}

func (se *SceneExplorerPanel) Draw() {

}
