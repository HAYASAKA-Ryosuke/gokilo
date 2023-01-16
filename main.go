package main

import (
	"flag"
	"fmt"
	ac "gokilo/autoCompletion"
	"gokilo/debug"
	"gokilo/lsp"
	rend "gokilo/render"
	"gokilo/snippet"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type EditorRow struct {
	text               string // データとしてもっておく文字列
	renderText         string // 表示用の文字列
	renderRowOffset    int
	renderColumnLength int
}

var (
	LSP            = &lsp.Lsp{}
	render         = []rend.Render{}
	autoCompletion = &ac.AutoCompletion{}
	page           = 0
)

func keyEnter(s tcell.Screen) {
	if autoCompletion.IsEnabled() {
		completions, index, _, _ := autoCompletion.GetCompletions(3)
		if len(completions) > 0 {
			render[page].EditorInsertText(completions[index])
		}
		autoCompletion.SetEnabled(false)
	} else {
		render[page].EditorInsertNewline(s)
	}
}

func keyUp() {
	if autoCompletion.IsEnabled() {
		autoCompletion.UpdateIndex(-1)
	} else {
		render[page].CursorMove("up")
	}
}

func keyDown() {
	if autoCompletion.IsEnabled() {
		autoCompletion.UpdateIndex(1)
	} else {
		render[page].CursorMove("down")
	}
}

func keyLeft() {
	render[page].CursorMove("left")
	autoCompletion.SetEnabled(false)
}

func keyRight() {
	render[page].CursorMove("right")
	autoCompletion.SetEnabled(false)
}

func keyCtrlP() {
	autoCompletion.SetEnabled(!autoCompletion.IsEnabled())
	currentRow, currentColumn := render[page].GetCurrentPosition()
	autoCompletion.UpdateAutoCompletion(render[page].GetCurrentFilePath(), LSP, currentRow, currentColumn)
}

func keyCtrlS() {
	fileSave()
}

func keyBackspace2(s tcell.Screen) {
	render[page].EditorDeleteChar(s)
}

func keyCtrlL(s tcell.Screen) {
	s.Sync()
}

func keyCtrlQ(s tcell.Screen) {
	quit(s)
}

func keyCtrlC() {

}

func keyEscape() {

}

func otherKey(ev *tcell.EventKey) {
	render[page].EditorInsertText(string(ev.Rune()))
}

func quit(s tcell.Screen) {
	s.Fini()
	os.Exit(0)
}

func editorProcessKeyPress(s tcell.Screen, ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		keyEscape()
	case tcell.KeyCtrlC:
		keyCtrlC()
	case tcell.KeyCtrlL:
		keyCtrlL(s)
	case tcell.KeyCtrlQ:
		keyCtrlQ(s)
	case tcell.KeyCtrlP:
		keyCtrlP()
	case tcell.KeyBackspace2:
		keyBackspace2(s)
	case tcell.KeyEnter:
		keyEnter(s)
	case tcell.KeyLeft:
		keyLeft()
	case tcell.KeyRight:
		keyRight()
	case tcell.KeyDown:
		keyDown()
	case tcell.KeyUp:
		keyUp()
	case tcell.KeyCtrlS:
		keyCtrlS()
	default:
		otherKey(ev)
	}
}

func fileSave() {
	var f *os.File
	f, _ = os.Create(render[page].GetCurrentFilePath())

	saveData := render[page].GetAllText()
	n, _ := f.WriteString(saveData)
	render[page].SetDebug(fmt.Sprintf("save: %d", n))
	f.Close()
}

func getArgs() string {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		path, _ := os.Getwd()
		return path + "/test.go"
	} else {
		return args[0]
	}
}

func refresh(s tcell.Screen) {
	render[page].EditorScroll(s)
	s.Clear()
	render[page].UpdateRenderRowAndColumn(s)
	render[page].EditorDrawRows(s)
	completions, index, selectedIndex, completionTotalCount := autoCompletion.GetCompletions(5)
	if completionTotalCount > 0 {
		render[page].SetDebug(fmt.Sprintf("%d,%d,%d,%d", len(completions), index, completionTotalCount, len(completions)*(selectedIndex)/completionTotalCount))
		renderRow, renderColumn := render[page].GetRenderPosition()
		snippet.DrawSnippet(s, renderColumn, renderRow+1, completions, index, len(completions)*(selectedIndex)/completionTotalCount)
	}
	render[page].UpdateShowCursor(s)
}

func main() {
	debug.LogConfig("./app.log")
	LSP = lsp.NewLsp("/home/hayasaka/go/bin/gopls")
	path, _ := os.Getwd()
	LSP.Init(path)

	currentFilePath := getArgs()

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
	ox, _ := -1, -1

	rr := rend.NewRender(
		s,
		"\n",
		" ",
		8,
		"dracula",
		"go",
		"debug",
		currentFilePath,
		0,
		0,
		0,
		0,
		"",
		0,
		7,
		0,
		1,
		false,
	)
	render = []rend.Render{rr}

	for {
		refresh(s)
		//s.ShowCursor(renderColumn-columnOffset, renderRow)
		// Update screen
		//s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			render[page].UpdateWindowSize(s)
		case *tcell.EventKey:
			editorProcessKeyPress(s, ev)
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := ev.Buttons()
			// Only process button events, not wheel events
			button &= tcell.ButtonMask(0xff)

			if button != tcell.ButtonNone && ox < 0 {
				ox, _ = x, y
			}
			switch ev.Buttons() {
			case tcell.ButtonNone:
				if ox >= 0 {
					ox, _ = -1, -1
				}
			}
		}
	}
}
