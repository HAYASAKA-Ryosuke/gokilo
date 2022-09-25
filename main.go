package main

import (
	"fmt"
	"gokilo/highlight"
	"gokilo/lsp"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type EditorRow struct {
	text               string // データとしてもっておく文字列
	renderText         string // 表示用の文字列
	renderRowOffset    int
	renderColumnLength int
}

var (
	TAB_SIZE               = 8
	SYNTAX_HIGHLIGHT_STYLE = "dracula"
	LANGUAGE               = "go"
	DEBUG                  = "debug"
	currentColumn          = 0
	currentRow             = 0
	renderColumn           = 0
	renderRow              = 0
	autoCompletionEnable   = false
	editorBuf              = ""
	rowOffset              = 0
	columnOffset           = 0
	editorRows             = []EditorRow{}
	defStyle               = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	windowSizeColumn       = 0
	windowSizeRow          = 0
)

func drawContent(s tcell.Screen, column, row int, text string, textColorStyle tcell.Style) (int, int) {

	//textColorStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.PaletteColor(1))
	for _, r := range []rune(text) {
		if windowSizeColumn <= column {
			row++
			column = 0
		}
		s.SetContent(column, row, r, nil, textColorStyle)
		column += runewidth.RuneWidth(r)
	}
	return column, row

}

func drawStatusBar(s tcell.Screen) {
	//DEBUG = fmt.Sprintf("cCol %d, cRow %d, rCol %d, rRow %d, rowCol %d, rowRow %d", currentColumn, currentRow, renderColumn, renderRow, editorRows[currentRow].renderColumnLength, editorRows[currentRow].renderRowOffset)
	drawContent(s, 0, windowSizeRow, fmt.Sprintf("status %d, %d, %d, %s", currentColumn, currentRow, renderColumn, DEBUG), defStyle)
}

func convertAnsiColorCodeFormatToInt(ansiColorCode string) (int, error) {
	result := strings.Replace(ansiColorCode, "\x1b[38;5;", "", -1)
	result = strings.Replace(result, "\x1b[48;5;", "", -1)
	result = strings.Replace(result, "m", "", -1)
	result = strings.Trim(result, " ")
	colorCode, err := strconv.Atoi(result)
	if err != nil {
		DEBUG = fmt.Sprintf("%s, %#v, %s", ansiColorCode, result, err)
	}
	return colorCode, err
}

func editorDrawRows(s tcell.Screen) {
	//for y := currentRow; y < windowSizeRow; y++ {
	//	drawContent(s, 0, y, strconv.Itoa(y+1))
	//}
	row := 0
	for fileRow := rowOffset; fileRow < windowSizeRow-1; fileRow++ {
		if len(editorRows) <= fileRow {
			break
		}
		renderTextList, _ := highlight.Highlight(editorRows[fileRow].renderText, LANGUAGE, SYNTAX_HIGHLIGHT_STYLE)
		column := 0
		for _, renderText := range renderTextList {
			textStyle := tcell.StyleDefault.Italic(renderText.Italic).Underline(renderText.Underline).Bold(renderText.Bold)
			if renderText.ForegroundColor != "" {
				foregroundColorCode, _ := convertAnsiColorCodeFormatToInt(renderText.ForegroundColor)
				textStyle = textStyle.Foreground(tcell.PaletteColor(foregroundColorCode))
			}
			if renderText.BackgroundColor != "" {
				backgroundColorCode, _ := convertAnsiColorCodeFormatToInt(renderText.BackgroundColor)
				textStyle = textStyle.Background(tcell.PaletteColor(backgroundColorCode))
			}
			column, row = drawContent(s, column, row, renderText.Text, textStyle)
		}
		row += editorRows[fileRow].renderRowOffset
		row++
	}
	drawStatusBar(s)
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

func getRenderStringCount(text string) int {
	count := 0
	for _, c := range []rune(text) {
		if len(string(c)) == 1 {
			count++
		} else {
			count += 2
		}
	}
	return count
}

func getStringCount(text string) int {
	return len([]rune(text))
}

func updateRenderRowAndColumn(s tcell.Screen) {
	rowText := []rune(editorRows[currentRow].renderText)

	// タブ1つはスペース8個分に相当するのでその分も調整してrenderColumnを決定する必要がある
	tabLength := strings.Count(editorRows[currentRow].text[:currentColumn], "\t")
	renderColumn = getRenderStringCount(string(rowText[:currentColumn])) + tabLength*TAB_SIZE - tabLength

	renderRow = 0
	for row := 0; row < currentRow; row++ {
		renderRow += editorRows[row].renderRowOffset + 1
	}

	tabLength = strings.Count(editorRows[currentRow].text, "\t")
	if getRenderStringCount(string(rowText))+tabLength*8-tabLength > windowSizeColumn {
		renderRow += int(renderColumn / windowSizeColumn)
		renderColumn = renderColumn % windowSizeColumn
	}
}

func editorRefreshScreen(s tcell.Screen) {
	editorScroll(s)
	s.Clear()
	updateRenderRowAndColumn(s)
	editorDrawRows(s)
}

func editorInsertText(row, column int, text string) {

	runes := []rune(editorRows[row].text)

	beforeText := string(runes[:column])
	afterText := string(runes[column:])
	editorRows[row].text = beforeText + text + afterText
	column++
	currentColumn = column
	editorUpdateRow(row)
}

func editorInsertRow(s tcell.Screen, row int, text string) {
	if row == len(editorRows)-1 {
		editorRows = append(editorRows, EditorRow{text, "", 0, 0})
	} else {
		beforeRows := editorRows[:row]
		afterRows := editorRows[row:]
		beforeRows = append(beforeRows, EditorRow{text, "", 0, 0})
		editorRows = append(beforeRows, afterRows...)

	}
	editorUpdateRow(currentRow)
}

func editorUpdateRow(row int) {
	editorRows[row].renderText = strings.Replace(editorRows[row].text, "\t", strings.Repeat(" ", TAB_SIZE), -1)
	editorRows[row].renderColumnLength = getRenderStringCount(editorRows[row].renderText)
	editorRows[row].renderRowOffset = int(editorRows[row].renderColumnLength / windowSizeColumn)
}

func editorInsertNewline(s tcell.Screen) {
	rowText := []rune(editorRows[currentRow].text)
	beforeText := rowText[currentColumn:]
	editorInsertRow(s, currentRow, string(beforeText))
	editorRows[currentRow].text = string(rowText[:currentColumn])
	editorUpdateRow(currentRow)
	currentRow++
	currentColumn = 0
	editorUpdateRow(currentRow)
}

func deleteRow() {
	currentRow--
	for i := currentRow; i < len(editorRows)-1; i++ {
		if currentRow == i {
			currentColumn = getRenderStringCount(editorRows[i].text)
			editorRows[i].text += editorRows[i+1].text
			editorRows[i].renderText += editorRows[i+1].renderText
		} else {
			editorRows[i] = editorRows[i+1]
		}
		editorUpdateRow(i)
	}

	// 最後の行を削除
	editorRows = editorRows[:len(editorRows)-1]
	for i := 0; i < len(editorRows); i++ {
		editorUpdateRow(i)
	}
}

func editorDeleteChar(s tcell.Screen) {
	// TODO: 0列目のときバックスペースを押下するとその行の文字列が一つ上の行の末尾に結合されるようにすること
	if currentColumn != 0 {
		runes := []rune(editorRows[currentRow].text)
		editorRows[currentRow].text = string(runes[:currentColumn-1]) + string(runes[currentColumn:])
		currentColumn--
		editorUpdateRow(currentRow)
	} else {
		if currentRow != 0 {
			deleteRow()
		}
	}
}

func keyUp() {
	if currentRow != 0 {
		currentRow--
		if getStringCount(editorRows[currentRow].text) < currentColumn {
			currentColumn = getStringCount(editorRows[currentRow].text)
		}
	}
}

func keyDown() {
	if len(editorRows) > currentRow+1 {
		currentRow++
		renderRow++
		if getStringCount(editorRows[currentRow].text) < currentColumn {
			currentColumn = getStringCount(editorRows[currentRow].text)
		}
	}

	if len(editorRows) < currentRow+1 {
		editorRows = append(editorRows, EditorRow{"", "", 0, 0})
	}
}

func keyLeft() {
	if currentColumn != 0 {
		currentColumn--
	} else if currentColumn == 0 && currentRow != 0 {
		currentRow--
		currentColumn = getStringCount(editorRows[currentRow].text)
		editorUpdateRow(currentRow)
	}
}

func keyRight() {
	if currentColumn < getStringCount(editorRows[currentRow].text) {
		currentColumn++
		editorUpdateRow(currentRow)
	} else if currentRow < len(editorRows)-1 {
		currentColumn = 0
		currentRow++
		editorUpdateRow(currentRow)
	}
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
	} else if ev.Key() == tcell.KeyBackspace2 {
		editorDeleteChar(s)
	} else if ev.Key() == tcell.KeyEnter {
		editorInsertNewline(s)
	} else if ev.Key() == tcell.KeyLeft {
		keyLeft()
	} else if ev.Key() == tcell.KeyRight {
		keyRight()
	} else if ev.Key() == tcell.KeyDown {
		keyDown()
	} else if ev.Key() == tcell.KeyUp {
		keyUp()
	} else {
		editorInsertText(currentRow, currentColumn, string(ev.Rune()))
	}
}

func initialize(s tcell.Screen) {
	editorRows = append(editorRows, EditorRow{"", "", 0, 0})
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
	windowSizeRow--

	initialize(s)

	for {

		editorRefreshScreen(s)
		s.ShowCursor(renderColumn, renderRow)
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			windowSizeColumn, windowSizeRow = getWindowSize(s)
			windowSizeRow--
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
