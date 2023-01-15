package snippet

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func DrawSnippet(s tcell.Screen, startColumn, startRow int, completionList []string, selectedIndex int, scrollBarPosition int) {

	columnWidth := 10
	for i := 0; i < len(completionList); i++ {
		if columnWidth < len(completionList[i]) {
			columnWidth = len(completionList[i])
		}
	}
	columnWidth += 4

	unselectedStyle := tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.ColorWheat)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite)
	style := unselectedStyle
	// Fill background
	for row := 0; row < len(completionList); row++ {
		if row == selectedIndex {
			style = unselectedStyle
		} else {
			style = selectedStyle
		}
		for col := 0; col < columnWidth; col++ {
			var text []rune
			if scrollBarPosition == row {
				text = []rune(fmt.Sprintf("%-*s█", columnWidth-1, completionList[row]))
			} else {
				text = []rune(fmt.Sprintf("%-*s ", columnWidth-1, completionList[row]))
			}
			if len(text) > col {
				s.SetContent(col+startColumn, row+startRow, text[col], nil, style)
			} else {
				s.SetContent(col+startColumn, row+startRow, ' ', nil, style)
			}
		}
	}
}
