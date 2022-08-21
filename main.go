package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type EditorRow struct {
	text       string // データとしてもっておく文字列
	renderText string // 表示用の文字列
}

var (
	DEBUG                = "debug"
	currentColumn        = 0
	currentRow           = 0
	renderColumn         = 0
	renderRow            = 0
	autoCompletionEnable = false
	editorBuf            = ""
	rowOffset            = 0
	columnOffset         = 0
	editorRows           = []EditorRow{EditorRow{"", ""}}
	defStyle             = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	windowSizeColumn     = 0
	windowSizeRow        = 0
)

func drawContent(s tcell.Screen, column, row int, text string) {
	for _, r := range []rune(text) {
		s.SetContent(column, row, r, nil, defStyle)
		column++
	}

}

func drawStatusBar(s tcell.Screen) {
	drawContent(s, 0, windowSizeRow-1, fmt.Sprintf("status %d, %d, %s", currentColumn, currentRow, DEBUG))
}

func editorDrawRows(s tcell.Screen) {
	//for y := currentRow; y < windowSizeRow; y++ {
	//	drawContent(s, 0, y, strconv.Itoa(y+1))
	//}
	drawContent(s, renderColumn, renderRow, editorRows[currentRow].renderText)
	drawStatusBar(s)
}

func editorAppendRow(c string) {
	editorRowLength := len(editorRows) - 1
	editorRows[editorRowLength].text += c
	updateRenderRow(editorRows[editorRowLength])
}

func updateRenderRow(row EditorRow) {
	row.renderText = row.text
}

func editorScroll(c tcell.Screen) {
	if currentRow < rowOffset {
		rowOffset = currentRow
	}
	if currentRow >= rowOffset+windowSizeRow {
		rowOffset = currentRow - windowSizeRow + 1
	}
}

func getWindowSize(s tcell.Screen) (int, int) {
	return s.Size()
}

func editorRefreshScreen(s tcell.Screen) {
	editorScroll(s)
	s.Clear()
}

func quit(s tcell.Screen) {
	s.Fini()
	os.Exit(0)
}

func editorProcessKeyPress(s tcell.Screen, ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
	} else if ev.Key() == tcell.KeyCtrlL {
		s.Sync()
	} else if ev.Key() == tcell.KeyCtrlQ {
		quit(s)
	} else if ev.Key() == tcell.KeyCtrlP {
	} else if ev.Key() == tcell.KeyEnter {
	} else if ev.Key() == tcell.KeyLeft {
		if currentColumn != 0 {
			currentColumn--
		} else if currentColumn == 0 && currentRow != 0 {
			currentColumn = 0
			currentRow--
		}
	} else if ev.Key() == tcell.KeyRight {
		if currentColumn != windowSizeColumn-1 {
			currentColumn++
		} else if currentColumn == windowSizeColumn-1 && currentRow != windowSizeRow-1 {
			currentColumn = 0
			currentRow++
		}
	} else if ev.Key() == tcell.KeyDown {
		if currentRow != windowSizeRow-1 {
			currentRow++
		}
	} else if ev.Key() == tcell.KeyUp {
		if currentRow != 0 {
			currentRow--
		}
	} else {
	}
}

func initialize(s tcell.Screen) {

}

func main() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	//drawBox(s, 5, 9, 32, 14, completionColorStyle, "Press C to reset")

	// Event loop
	ox, oy := -1, -1
	fmt.Println(oy)

	windowSizeColumn, windowSizeRow = getWindowSize(s)

	for {

		editorRefreshScreen(s)
		s.ShowCursor(currentColumn, currentRow)
		editorDrawRows(s)
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			windowSizeColumn, windowSizeRow = getWindowSize(s)
		case *tcell.EventKey:
			editorProcessKeyPress(s, ev)
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := ev.Buttons()
			// Only process button events, not wheel events
			button &= tcell.ButtonMask(0xff)

			if button != tcell.ButtonNone && ox < 0 {
				ox, oy = x, y
			}
			switch ev.Buttons() {
			case tcell.ButtonNone:
				if ox >= 0 {
					ox, oy = -1, -1
				}
			}
		}
	}
}
