package render

import (
	"fmt"
	"gokilo/highlight"
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

type Render struct {
	newlineChar           string
	tabChar               string
	tabSize               int
	syntaxHighlightStyle  string
	language              string
	debug                 string
	currentFilePath       string
	currentColumn         int
	currentRow            int
	renderColumn          int
	renderRow             int
	editorBuf             string
	rowOffset             int
	rowNumberColumnOffset int
	columnOffset          int
	editorRows            []EditorRow
	windowSizeColumn      int
	windowSizeRow         int
	statusBarOffset       int
	wordWrap              bool // 折り返すか
}

//var (
//	NEWLINE_CHAR           = "\n"
//	TAB_CHAR               = " "
//	TAB_SIZE               = 8
//	SYNTAX_HIGHLIGHT_STYLE = "dracula"
//	LANGUAGE               = "go"
//	DEBUG                  = "debug"
//	currentFilePath        = ""
//	currentColumn          = 0
//	currentRow             = 0
//	renderColumn           = 0
//	renderRow              = 0
//	editorBuf              = ""
//	rowOffset              = 0
//	rowNumberColumnOffset  = 7
//	columnOffset           = 0
//	editorRows             = []EditorRow{}

//	windowSizeColumn       = 0
//	windowSizeRow          = 0
//	STATUS_BAR_OFFSET      = 1
//	WORD_WRAP              = false // 折り返すか
//	LSP                    = &lsp.Lsp{}
//)

func NewRender(
	s tcell.Screen,
	newlineChar string,
	tabChar string,
	tabSize int,
	syntaxHighlightStyle string,
	language string,
	debug string,
	currentFilePath string,
	currentColumn int,
	currentRow int,
	renderColumn int,
	renderRow int,
	editorBuf string,
	rowOffset int,
	rowNumberColumnOffset int,
	columnOffset int,
	statusBarOffset int,
	wordWrap bool,
) Render {
	editorRows := []EditorRow{}
	editorRows = append(editorRows, EditorRow{"", "", 0, 0})
	render := Render{
		newlineChar:           newlineChar,
		tabChar:               tabChar,
		tabSize:               tabSize,
		syntaxHighlightStyle:  syntaxHighlightStyle,
		language:              language,
		debug:                 debug,
		currentFilePath:       currentFilePath,
		currentColumn:         currentColumn,
		currentRow:            currentRow,
		renderColumn:          renderColumn,
		renderRow:             renderRow,
		editorBuf:             editorBuf,
		rowOffset:             rowOffset,
		rowNumberColumnOffset: rowNumberColumnOffset,
		editorRows:            editorRows,
		columnOffset:          columnOffset,
		statusBarOffset:       statusBarOffset,
		wordWrap:              wordWrap,
	}
	render.UpdateWindowSize(s)
	return render
}

func (render *Render) SetDebug(text string) {
	render.debug = text
}

func (render *Render) UpdateShowCursor(s tcell.Screen) {
	s.ShowCursor(render.renderColumn-render.columnOffset, render.renderRow)
	s.Show()
}

func (render *Render) UpdateWindowSize(s tcell.Screen) {
	render.windowSizeColumn, render.windowSizeRow = render.GetWindowSize(s)
	render.windowSizeRow--
}

func (render *Render) drawContent(s tcell.Screen, column, row int, text string, textColorStyle tcell.Style) (int, int) {

	//textColorStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.PaletteColor(1))
	for _, r := range []rune(text) {
		if render.wordWrap {
			if render.windowSizeColumn <= column {
				row++
				column = 0
			}
		}
		s.SetContent(column, row, r, nil, textColorStyle)
		column += runewidth.RuneWidth(r)
	}
	return column, row

}

func (render *Render) DrawStatusBar(s tcell.Screen) {
	//DEBUG = fmt.Sprintf("cCol %d, cRow %d, rCol %d, rRow %d, rowCol %d, rowRow %d, rowOffset %d", currentColumn, currentRow, renderColumn, renderRow, editorRows[currentRow].renderColumnLength, editorRows[currentRow].renderRowOffset, rowOffset)
	style := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorReset)
	text := fmt.Sprintf("status %d, %d, %d, %d, %s", render.currentColumn, render.currentRow, render.renderColumn, render.renderRow, render.debug)
	render.drawContent(s, 0, render.windowSizeRow, text+strings.Repeat(" ", render.windowSizeColumn-len(text)), style)
}

func (render *Render) ConvertAnsiColorCodeFormatToInt(ansiColorCode string) (int, error) {
	result := strings.Replace(ansiColorCode, "\x1b[38;5;", "", -1)
	result = strings.Replace(result, "\x1b[48;5;", "", -1)
	result = strings.Replace(result, "m", "", -1)
	result = strings.Trim(result, " ")
	colorCode, err := strconv.Atoi(result)
	if err != nil {
		render.debug = fmt.Sprintf("%s, %#v, %s", ansiColorCode, result, err)
	}
	return colorCode, err
}

func (render *Render) EditorDrawRows(s tcell.Screen) {
	//for y := currentRow; y < windowSizeRow; y++ {
	//	drawContent(s, 0, y, strconv.Itoa(y+1))
	//}
	row := 0
	for i := 0; i < render.windowSizeRow; i++ {
		if len(render.editorRows) <= i+render.rowOffset {
			break
		}
		renderText := render.editorRows[i+render.rowOffset].renderText
		if !render.wordWrap {
			if len(renderText) >= render.columnOffset {
				renderText = renderText[render.columnOffset:]
			}
		}
		renderTextList, _ := highlight.Highlight(renderText, render.language, render.syntaxHighlightStyle)
		column := 0
		for _, renderText := range renderTextList {
			textStyle := tcell.StyleDefault.Italic(renderText.Italic).Underline(renderText.Underline).Bold(renderText.Bold)
			if renderText.ForegroundColor != "" {
				foregroundColorCode, _ := render.ConvertAnsiColorCodeFormatToInt(renderText.ForegroundColor)
				textStyle = textStyle.Foreground(tcell.PaletteColor(foregroundColorCode))
			}
			if renderText.BackgroundColor != "" {
				backgroundColorCode, _ := render.ConvertAnsiColorCodeFormatToInt(renderText.BackgroundColor)
				textStyle = textStyle.Background(tcell.PaletteColor(backgroundColorCode))
			}
			if render.wordWrap {
				column, row = render.drawContent(s, column%render.windowSizeColumn, row, renderText.Text, textStyle)
			} else {
				column, row = render.drawContent(s, column, row, renderText.Text, textStyle)
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
	render.DrawStatusBar(s)
}

func (render *Render) EditorScroll(c tcell.Screen) {
	tabLength := strings.Count(string([]rune(render.editorRows[render.currentRow].text)[:render.currentColumn]), "\t")
	render.renderColumn = render.GetRenderStringCount(string([]rune(render.editorRows[render.currentRow].text)[:render.currentColumn])) + tabLength*render.tabSize - tabLength

	if render.currentRow < render.rowOffset {
		render.rowOffset = render.currentRow
	}
	if render.currentRow >= render.rowOffset+render.windowSizeRow {
		render.rowOffset = render.currentRow - render.windowSizeRow + 1 + render.editorRows[render.currentRow].renderRowOffset
	}
	if render.renderColumn < render.columnOffset {
		render.columnOffset = render.renderColumn
	}
	if render.renderColumn >= render.columnOffset+render.windowSizeColumn {
		render.columnOffset = render.renderColumn - render.windowSizeColumn + 1
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

	if render.renderRow == render.windowSizeRow-1 {
		render.rowOffset = 0
		for row := render.currentRow; row >= 0; row-- {
			render.rowOffset += render.editorRows[row].renderRowOffset + 1
		}
		render.rowOffset -= render.windowSizeRow
		if render.rowOffset < 0 {
			render.rowOffset = 0
		}
	}
}

func (render *Render) GetNewlineCHar() string {
	return render.newlineChar
}

func (render *Render) GetAllText() string {
	result := ""
	for i := 0; i < len(render.editorRows); i++ {
		result += render.editorRows[i].text
		if i-1 != len(render.editorRows) {
			result += render.newlineChar
		}
	}
	return result
}

func (render *Render) GetEditorRows() []EditorRow {
	return render.editorRows
}

func (render *Render) GetCurrentFilePath() string {
	return render.currentFilePath
}

func (render *Render) GetRenderPosition() (int, int) {
	return render.renderRow, render.renderColumn
}

func (render *Render) GetCurrentPosition() (int, int) {
	return render.currentRow, render.currentColumn
}

func (render *Render) GetWindowSize(s tcell.Screen) (int, int) {
	return s.Size()
}

func (render *Render) GetRenderStringCount(text string) int {
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

func (render *Render) GetStringCount(text string) int {
	return len([]rune(text))
}

func (render *Render) UpdateRenderRowAndColumn(s tcell.Screen) {
	rowText := []rune(render.editorRows[render.currentRow].renderText)

	// タブ1つはスペース8個分に相当するのでその分も調整してrenderColumnを決定する必要がある
	tabLength := strings.Count(string([]rune(render.editorRows[render.currentRow].text)[:render.currentColumn]), "\t")
	render.renderColumn = render.GetRenderStringCount(string([]rune(render.editorRows[render.currentRow].text)[:render.currentColumn])) + tabLength*render.tabSize - tabLength

	render.renderRow = 0
	for row := render.currentRow - 1; row >= render.rowOffset; row-- {
		render.renderRow += render.editorRows[row].renderRowOffset + 1
	}

	tabLength = strings.Count(render.editorRows[render.currentRow].text, "\t")
	if render.GetRenderStringCount(string(rowText))+tabLength*render.tabSize-tabLength > render.windowSizeColumn {
		if render.wordWrap {
			render.renderRow += int(render.renderColumn / render.windowSizeColumn)
			render.renderColumn = render.renderColumn % render.windowSizeColumn
		}
	}

	if render.renderRow > render.windowSizeRow-render.statusBarOffset {
		render.renderRow = render.windowSizeRow - render.statusBarOffset
	}
}

func (render *Render) EditorInsertText(text string) {

	runes := []rune(render.editorRows[render.currentRow].text)

	beforeText := string(runes[:render.currentColumn])
	afterText := string(runes[render.currentColumn:])
	render.editorRows[render.currentRow].text = beforeText + text + afterText
	render.currentColumn += len([]rune(text))
	render.EditorUpdateRow(render.currentRow)
}

func (render *Render) EditorInsertRow(s tcell.Screen, row int, text string) {
	if row == len(render.editorRows)-1 {
		render.editorRows = append(render.editorRows, EditorRow{text, "", 0, 0})
	} else {
		beforeRows := render.editorRows[:row]
		afterRows := render.editorRows[row:]
		beforeRows = append(beforeRows, EditorRow{text, "", 0, 0})
		render.editorRows = append(beforeRows, afterRows...)

	}
	render.EditorUpdateRow(render.currentRow)
}

func (render *Render) EditorUpdateRow(row int) {
	render.editorRows[row].renderText = strings.Replace(render.editorRows[row].text, "\t", strings.Repeat(render.tabChar, render.tabSize), -1)
	render.editorRows[row].renderColumnLength = render.GetRenderStringCount(render.editorRows[row].renderText)
	if render.wordWrap {
		render.editorRows[row].renderRowOffset = int(render.editorRows[row].renderColumnLength / render.windowSizeColumn)
	}
}

func (render *Render) EditorInsertNewline(s tcell.Screen) {
	rowText := []rune(render.editorRows[render.currentRow].text)
	beforeText := rowText[render.currentColumn:]
	render.EditorInsertRow(s, render.currentRow, string(beforeText))
	render.editorRows[render.currentRow].text = string(rowText[:render.currentColumn])
	render.EditorUpdateRow(render.currentRow)
	render.currentRow++
	render.currentColumn = 0
	render.EditorUpdateRow(render.currentRow)
}

func (render *Render) DeleteRow() {
	render.currentRow--
	for i := render.currentRow; i < len(render.editorRows)-1; i++ {
		if render.currentRow == i {
			render.currentColumn = render.GetStringCount(render.editorRows[i].text)
			render.editorRows[i].text += render.editorRows[i+1].text
			render.editorRows[i].renderText += render.editorRows[i+1].renderText
		} else {
			render.editorRows[i] = render.editorRows[i+1]
		}
		render.EditorUpdateRow(i)
	}

	// 最後の行を削除
	render.editorRows = render.editorRows[:len(render.editorRows)-1]
	for i := 0; i < len(render.editorRows); i++ {
		render.EditorUpdateRow(i)
	}
}

func (render *Render) EditorDeleteChar(s tcell.Screen) {
	if render.currentColumn != 0 {
		runes := []rune(render.editorRows[render.currentRow].text)
		render.editorRows[render.currentRow].text = string(runes[:render.currentColumn-1]) + string(runes[render.currentColumn:])
		render.currentColumn--
		render.EditorUpdateRow(render.currentRow)
	} else {
		if render.currentRow != 0 {
			render.DeleteRow()
		}
	}
}

func (render *Render) CursorJump(s tcell.Screen, row int, col int) {
	render.currentRow = row
	render.currentColumn = col
	render.UpdateRenderRowAndColumn(s)

}

func (render *Render) CursorMove(to string) {
	switch to {
	case "up":
		if render.currentRow != 0 {
			render.currentRow--
			render.renderRow--
			if render.GetStringCount(render.editorRows[render.currentRow].text) < render.currentColumn {
				render.currentColumn = render.GetStringCount(render.editorRows[render.currentRow].text)
			}
		}
	case "down":
		if len(render.editorRows) > render.currentRow+1 {
			render.currentRow++
			render.renderRow++
			if render.renderRow > render.windowSizeRow-1 {
				render.renderRow = render.windowSizeRow - 1
			}
			if render.GetStringCount(render.editorRows[render.currentRow].text) < render.currentColumn {
				render.currentColumn = render.GetStringCount(render.editorRows[render.currentRow].text)
			}
		} else if len(render.editorRows) == render.currentRow+1 {
			render.currentColumn = render.GetStringCount(render.editorRows[render.currentRow].text)
		}

		if len(render.editorRows) < render.currentRow+1 {
			render.editorRows = append(render.editorRows, EditorRow{"", "", 0, 0})
		}
	case "right":
		if render.currentColumn < render.GetStringCount(render.editorRows[render.currentRow].text) {
			render.currentColumn++
			render.EditorUpdateRow(render.currentRow)
		} else if render.currentRow < len(render.editorRows)-1 {
			render.currentColumn = 0
			render.currentRow++
			render.EditorUpdateRow(render.currentRow)
		}
	case "left":
		if render.currentColumn != 0 {
			render.currentColumn--
		} else if render.currentColumn == 0 && render.currentRow != 0 {
			render.currentRow--
			render.currentColumn = render.GetStringCount(render.editorRows[render.currentRow].text)
			render.EditorUpdateRow(render.currentRow)
		}
	}
}
