package main

import (
	"flag"
	"fmt"
	"gokilo/debug"
	"gokilo/highlight"
	"gokilo/snippet"
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

type CompletionInfo struct {
	icon string
	text string
}

type AutoCompletion struct {
	completionList []string
	selectedIndex  int
}

var (
	autoCompletionEnable   = false
	autoCompletion         = AutoCompletion{completionList: []string{"Println", "Printf", "Append", "Appendf", "foo", "hoge", "ham", "egg", "spam"}, selectedIndex: 0}
	NEWLINE_CHAR           = "\n"
	TAB_CHAR               = " "
	TAB_SIZE               = 8
	SYNTAX_HIGHLIGHT_STYLE = "dracula"
	LANGUAGE               = "go"
	DEBUG                  = "debug"
	filePath               = ""
	currentColumn          = 0
	currentRow             = 0
	renderColumn           = 0
	renderRow              = 0
	editorBuf              = ""
	rowOffset              = 0
	rowNumberColumnOffset  = 7
	columnOffset           = 0
	editorRows             = []EditorRow{}
	defStyle               = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	windowSizeColumn       = 0
	windowSizeRow          = 0
	STATUS_BAR_OFFSET      = 1
	WORD_WRAP              = false // 折り返すか
)

func drawContent(s tcell.Screen, column, row int, text string, textColorStyle tcell.Style) (int, int) {

	//textColorStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.PaletteColor(1))
	for _, r := range []rune(text) {
		if WORD_WRAP {
			if windowSizeColumn <= column {
				row++
				column = 0
			}
		}
		s.SetContent(column, row, r, nil, textColorStyle)
		column += runewidth.RuneWidth(r)
	}
	return column, row

}

func drawStatusBar(s tcell.Screen) {
	//DEBUG = fmt.Sprintf("cCol %d, cRow %d, rCol %d, rRow %d, rowCol %d, rowRow %d, rowOffset %d", currentColumn, currentRow, renderColumn, renderRow, editorRows[currentRow].renderColumnLength, editorRows[currentRow].renderRowOffset, rowOffset)
	style := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorReset)
	text := fmt.Sprintf("status %d, %d, %d, %d, %s", currentColumn, currentRow, renderColumn, renderRow, DEBUG)
	drawContent(s, 0, windowSizeRow, text+strings.Repeat(" ", windowSizeColumn-len(text)), style)
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
	for i := 0; i < windowSizeRow; i++ {
		if len(editorRows) <= i+rowOffset {
			break
		}
		renderText := editorRows[i+rowOffset].renderText
		if !WORD_WRAP {
			if len(renderText) >= columnOffset {
				renderText = renderText[columnOffset:]
			}
		}
		renderTextList, _ := highlight.Highlight(renderText, LANGUAGE, SYNTAX_HIGHLIGHT_STYLE)
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
			if WORD_WRAP {
				column, row = drawContent(s, column%windowSizeColumn, row, renderText.Text, textStyle)
			} else {
				column, row = drawContent(s, column, row, renderText.Text, textStyle)
			}
		}
		row++
	}
	//if len(editorRows) < windowSizeRow {
	//	for row := 0; row < len(editorRows); row++ {
	//		drawContent(s, 0, row, fmt.Sprintf("%d", row+rowOffset+1), defStyle)
	//	}
	//} else if rowOffset == windowSizeRow-1 || rowOffset == 0 {
	//	for row := 0; row < windowSizeRow; row++ {
	//		drawContent(s, 0, row, fmt.Sprintf("%d", row+rowOffset+1), defStyle)
	//	}
	//}
	drawStatusBar(s)
}

func editorScroll(c tcell.Screen) {

	tabLength := strings.Count(string([]rune(editorRows[currentRow].text)[:currentColumn]), "\t")
	renderColumn = getRenderStringCount(string([]rune(editorRows[currentRow].text)[:currentColumn])) + tabLength*TAB_SIZE - tabLength

	if currentRow < rowOffset {
		rowOffset = currentRow
	}
	if currentRow >= rowOffset+windowSizeRow {
		rowOffset = currentRow - windowSizeRow + 1 + editorRows[currentRow].renderRowOffset
	}
	if renderColumn < columnOffset {
		columnOffset = renderColumn
	}
	if renderColumn >= columnOffset+windowSizeColumn {
		columnOffset = renderColumn - windowSizeColumn + 1
	}
	////if renderRow == windowSizeRow-1 && ((editorRows[currentRow].renderRowOffset+1)*windowSizeColumn >= editorRows[currentRow].renderColumnLength && (editorRows[currentRow].renderRowOffset)*windowSizeColumn < editorRows[currentRow].renderColumnLength) {
	//if renderRow == windowSizeRow-1 {
	//	rowOffset = 0
	//	for i := currentRow; i >= 0; i-- {
	//		rowOffset += editorRows[i].renderRowOffset + 1
	//	}
	//	rowOffset -= windowSizeRow
	//	//rowOffset = currentRow - windowSizeRow + 1 + editorRows[currentRow].renderRowOffset
	//}

	if renderRow == windowSizeRow-1 {
		rowOffset = 0
		for row := currentRow; row >= 0; row-- {
			rowOffset += editorRows[row].renderRowOffset + 1
		}
		rowOffset -= windowSizeRow
		if rowOffset < 0 {
			rowOffset = 0
		}
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
	tabLength := strings.Count(string([]rune(editorRows[currentRow].text)[:currentColumn]), "\t")
	renderColumn = getRenderStringCount(string([]rune(editorRows[currentRow].text)[:currentColumn])) + tabLength*TAB_SIZE - tabLength

	renderRow = 0
	for row := currentRow - 1; row >= rowOffset; row-- {
		renderRow += editorRows[row].renderRowOffset + 1
	}

	tabLength = strings.Count(editorRows[currentRow].text, "\t")
	if getRenderStringCount(string(rowText))+tabLength*TAB_SIZE-tabLength > windowSizeColumn {
		if WORD_WRAP {
			renderRow += int(renderColumn / windowSizeColumn)
			renderColumn = renderColumn % windowSizeColumn
		}
	}

	if renderRow > windowSizeRow-STATUS_BAR_OFFSET {
		renderRow = windowSizeRow - STATUS_BAR_OFFSET
	}
}

func editorRefreshScreen(s tcell.Screen) {
	editorScroll(s)
	s.Clear()
	updateRenderRowAndColumn(s)
	editorDrawRows(s)
	if autoCompletionEnable {
		showAutoCompletion(s)
	}
}

func editorInsertText(row, column int, text string) {

	runes := []rune(editorRows[row].text)

	beforeText := string(runes[:column])
	afterText := string(runes[column:])
	editorRows[row].text = beforeText + text + afterText
	column += len([]rune(text))
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
	editorRows[row].renderText = strings.Replace(editorRows[row].text, "\t", strings.Repeat(TAB_CHAR, TAB_SIZE), -1)
	editorRows[row].renderColumnLength = getRenderStringCount(editorRows[row].renderText)
	if WORD_WRAP {
		editorRows[row].renderRowOffset = int(editorRows[row].renderColumnLength / windowSizeColumn)
	}
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
			currentColumn = getStringCount(editorRows[i].text)
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
		renderRow--
		if getStringCount(editorRows[currentRow].text) < currentColumn {
			currentColumn = getStringCount(editorRows[currentRow].text)
		}
	}

	if autoCompletionEnable {
		if autoCompletion.selectedIndex == 0 {
			autoCompletion.selectedIndex = len(autoCompletion.completionList) - 1
		} else {
			autoCompletion.selectedIndex = (autoCompletion.selectedIndex - 1) % len(autoCompletion.completionList)
		}
	}
}

func keyDown() {
	if len(editorRows) > currentRow+1 {
		currentRow++
		renderRow++
		if renderRow > windowSizeRow-1 {
			renderRow = windowSizeRow - 1
		}
		if getStringCount(editorRows[currentRow].text) < currentColumn {
			currentColumn = getStringCount(editorRows[currentRow].text)
		}
	}

	if len(editorRows) < currentRow+1 {
		editorRows = append(editorRows, EditorRow{"", "", 0, 0})
	}

	if autoCompletionEnable {
		if len(autoCompletion.completionList) != 0 {
			autoCompletion.selectedIndex = (autoCompletion.selectedIndex + 1) % len(autoCompletion.completionList)
		}
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
	autoCompletionEnable = false
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
	autoCompletionEnable = false
}

func showAutoCompletion(s tcell.Screen) {
	snippet.DrawSnippet(s, renderColumn, renderRow, autoCompletion.completionList, autoCompletion.selectedIndex)
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
		autoCompletionEnable = !autoCompletionEnable
	} else if ev.Key() == tcell.KeyBackspace2 {
		editorDeleteChar(s)
	} else if ev.Key() == tcell.KeyEnter {
		if autoCompletionEnable {
			autoCompletionEnable = false
			editorInsertText(currentRow, currentColumn, autoCompletion.completionList[autoCompletion.selectedIndex])
		} else {
			editorInsertNewline(s)
		}
	} else if ev.Key() == tcell.KeyLeft {
		keyLeft()
	} else if ev.Key() == tcell.KeyRight {
		keyRight()
	} else if ev.Key() == tcell.KeyDown {
		keyDown()
	} else if ev.Key() == tcell.KeyUp {
		keyUp()
	} else if ev.Key() == tcell.KeyCtrlS {
		fileSave()
	} else {
		editorInsertText(currentRow, currentColumn, string(ev.Rune()))
	}
}

func fileSave() {
	var f *os.File
	f, _ = os.Create(filePath)

	saveData := ""
	for i := 0; i < len(editorRows); i++ {
		saveData += editorRows[i].text
		if i-1 != len(editorRows) {
			saveData += NEWLINE_CHAR
		}
	}
	n, _ := f.WriteString(saveData)
	DEBUG = fmt.Sprintf("save: %d", n)
	f.Close()
}

func initialize(s tcell.Screen) {
	editorRows = append(editorRows, EditorRow{"", "", 0, 0})
}

func getArgs() string {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return "./test.c"
	} else {
		return args[0]
	}
}

func main() {

	debug.LogConfig("./app.log")
	log.Println("hello")

	filePath = getArgs()

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
		s.ShowCursor(renderColumn-columnOffset, renderRow)
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
