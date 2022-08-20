package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type EditorRow struct {
	text string
}

var (
	windowX              = 0
	windowY              = 0
	currentX             = 0
	currentY             = 0
	autoCompletionEnable = false
	editorBuf            = ""
	editorRow            = []EditorRow{EditorRow{""}}
	defStyle             = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
)

func editorDrawRows(s tcell.Screen) {
	for y := 0; y < 24; y++ {
		for _, r := range []rune("~") {
			s.SetContent(0, y, r, nil, defStyle)
		}
	}
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	windowWidth, _ := s.Size()
	for _, r := range []rune(text) {
		if windowWidth < x1 {
			col++
			windowY++
		}
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func editorRefreshScreen(s tcell.Screen) {
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
	} else if ev.Key() == tcell.KeyRight {
	} else if ev.Key() == tcell.KeyDown {
	} else if ev.Key() == tcell.KeyUp {
	} else {
	}

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

	for {

		s.ShowCursor(windowX, windowY)
		editorDrawRows(s)
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
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
