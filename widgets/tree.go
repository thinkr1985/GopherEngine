package widgets

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type TreeNode struct {
	Label      string
	Children   []*TreeNode
	IsExpanded bool
}

type TreeWidget struct {
	Bounds      rl.Rectangle
	Nodes       []*TreeNode
	FontSize    int32
	TextColor   rl.Color
	Indentation float32
}

func NewTreeWidget(xPos, yPos, width, height int, fontSize int32) *TreeWidget {
	return &TreeWidget{
		Bounds:      rl.NewRectangle(float32(xPos), float32(yPos), float32(width), float32(height)),
		FontSize:    fontSize,
		TextColor:   rl.White,
		Indentation: 20,
	}
}

func (tw *TreeWidget) AddNode(label string, parent *TreeNode) *TreeNode {
	node := &TreeNode{Label: label, IsExpanded: false}
	if parent == nil {
		tw.Nodes = append(tw.Nodes, node)
	} else {
		parent.Children = append(parent.Children, node)
	}
	return node
}
func (tw *TreeWidget) Draw() {
	// Draw each node in the tree
	yPos := tw.Bounds.Y + 10
	for _, node := range tw.Nodes {
		yPos = tw.drawNode(node, 0, yPos)
	}
}

func (tw *TreeWidget) drawNode(node *TreeNode, level int, yPos float32) float32 {
	// Adjust the position for each level (indentation)
	xPos := tw.Bounds.X + float32(level)*tw.Indentation
	lineHeight := float32(tw.FontSize + 5)
	toggleSize := float32(tw.FontSize)

	// Draw the toggle button (expand/collapse)
	toggleWidth := toggleSize
	toggleHeight := toggleSize
	togglePos := rl.NewRectangle(xPos, yPos, toggleWidth, toggleHeight)

	// Draw vertical line from parent to child if not root level
	if level > 0 {
		rl.DrawLineEx(
			rl.Vector2{X: xPos - tw.Indentation/2, Y: yPos - lineHeight/2},
			rl.Vector2{X: xPos - tw.Indentation/2, Y: yPos + toggleHeight},
			1,
			rl.Gray,
		)
	}

	// Toggle button interaction
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), togglePos) {
		// Change color when hovered
		rl.DrawRectangleRec(togglePos, rl.DarkGray)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			node.IsExpanded = !node.IsExpanded
		}
	} else {
		// Default color when not hovered
		rl.DrawRectangleRec(togglePos, rl.Gray)
	}

	// Draw the "plus/minus" sign to indicate expand/collapse
	if len(node.Children) > 0 {
		if node.IsExpanded {
			rl.DrawText("-", int32(xPos+3), int32(yPos+3), tw.FontSize, rl.Black)
		} else {
			rl.DrawText("+", int32(xPos+3), int32(yPos+3), tw.FontSize, rl.Black)
		}
	} else {
		// Draw a small dot for leaf nodes
		rl.DrawCircle(int32(xPos+toggleWidth/2), int32(yPos+toggleHeight/2), 2, rl.Black)
	}

	// Draw the label for the node
	rl.DrawText(node.Label, int32(xPos+toggleWidth+5), int32(yPos), tw.FontSize, tw.TextColor)

	// Draw horizontal line connecting the toggle to the text
	if level > 0 {
		rl.DrawLineEx(
			rl.Vector2{X: xPos - tw.Indentation/2, Y: yPos + toggleHeight/2},
			rl.Vector2{X: xPos, Y: yPos + toggleHeight/2},
			1,
			rl.Gray,
		)
	}

	// Calculate the next Y position
	nextYPos := yPos + lineHeight

	// If the node is expanded, recursively draw its children
	if node.IsExpanded && len(node.Children) > 0 {
		// Draw vertical line that will connect all children
		firstChildY := nextYPos + lineHeight/2
		lastChildY := nextYPos
		for _, childNode := range node.Children {
			lastChildY = tw.drawNode(childNode, level+1, lastChildY)
		}

		// Draw the vertical line after we know the last child's position
		if len(node.Children) > 0 {
			rl.DrawLineEx(
				rl.Vector2{X: xPos + toggleWidth/2, Y: firstChildY},
				rl.Vector2{X: xPos + toggleWidth/2, Y: lastChildY - lineHeight/2},
				1,
				rl.Gray,
			)
		}

		nextYPos = lastChildY
	}

	return nextYPos
}

func (tw *TreeWidget) Update() {
	// This function can be used to update the state of the widget.
	// For now, this handles the interaction and does the drawing.
}
