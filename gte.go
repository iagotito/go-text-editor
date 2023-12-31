package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var logger *log.Logger

var ROWS, COLS int
var offsetRow, offsetCol int

var sourceFile string
var textBuffer = [][]rune{}

var modified bool
var saved bool

const (
	normalMode = iota
	insertMode
)
var mode int = 0

type cursorPos struct {
	Row, Col int

	// ColBufer is used to store the position the cursor was in the the original
	// line when moving vertically to restore the original column. For exemple,
	// when moving from position 10 to a line that only have 5 characters, the
	// cursor will be at position 5, but when move again, it should go back to
	// position 10, unless there is a horizontal move.
	ColBuffer int
}
var cursor = &cursorPos{0, 0, -1}

func init() {
 logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	log.SetOutput(logFile)
}

func main() {
	if len(os.Args) == 1 {
		log.Println("No source file provided.")
		fmt.Println("No source file provided.")
		return
	}
	sourceFile = os.Args[1]

	readFile(sourceFile)
	runEditor()
}

func max(a, b int) int {
	if a > b { return a }
	return b
}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	row := x
	col := y
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
	}
}

func readFile(filename string) {
	sourceFile = filename

	file, err := os.Open(filename)
	if err != nil {
		textBuffer = append(textBuffer, []rune{})
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lineBuffer := make([]rune, len(line))

		for pos, char := range line {
			lineBuffer[pos] = char
		}

		textBuffer = append(textBuffer, lineBuffer)
	}

	if len(textBuffer) == 0 {
		textBuffer = append(textBuffer, []rune{})
	}
}

func writeFile(filename string) {
  file, err := os.Create(filename)
  if err != nil { log.Printf("%v", err) }
  defer file.Close()

  writer := bufio.NewWriter(file)
  for row, line := range textBuffer {
      new_line := "\n"
      if row == len(textBuffer) { new_line = "" }
      write_line := string(line) + new_line
      _, err = writer.WriteString(write_line)
      if err != nil { fmt.Println("Error:", err) }
  }
	writer.Flush();
}

func scrollTextBuffer() {
  if cursor.Row < offsetRow { offsetRow = cursor.Row }
  if cursor.Col < offsetCol { offsetCol = cursor.Col }
  if cursor.Row >= offsetRow + ROWS { offsetRow = cursor.Row - ROWS+1 }
  if cursor.Col >= offsetCol + COLS { offsetCol = cursor.Col - COLS+1 }
}

func displayTextBuffer(s tcell.Screen) {
	var row, col int
	for row = 0; row < ROWS; row++ {
		rowPos := row + offsetRow
		for col = 0; col < COLS; col++ {
			colPos := col + offsetCol

			if rowPos >= 0 && rowPos < len(textBuffer) && colPos < len(textBuffer[rowPos]) {
				if textBuffer[rowPos][colPos] != '\t' {
					s.SetContent(col, row, textBuffer[rowPos][colPos], nil, tcell.StyleDefault)
				} else {
					s.SetContent(col, row, ' ', nil, tcell.StyleDefault.Background(tcell.ColorLightGreen))
				}
			} else if row + offsetRow >= len(textBuffer) {
					s.SetContent(0, row, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorBlue))
			}

			if rowPos < len(textBuffer) && colPos == len(textBuffer[rowPos]) {
				s.SetContent(col, row, '\n', nil, tcell.StyleDefault)
			}
		}
	}
}

func displayStatusBar(s tcell.Screen) {
	var modeStatus string
	var displayFilename string
	var fileStatus string
	var cursorStatus string
	var statusBarColor tcell.Color

	switch mode {
	case normalMode:
		modeStatus = " [ NORMAL ] "
		statusBarColor = tcell.ColorLightBlue
	case insertMode:
		modeStatus = " [ INSERT ] "
		statusBarColor = tcell.ColorOrange
	}

	if saved {
		fileStatus = " (saved)"
		if mode != insertMode {
			statusBarColor = tcell.ColorLightGreen
		}
	} else if modified {
		fileStatus = " (modified)"
	}

	displayFilename = sourceFile
	cursorStatus = fmt.Sprintf("%d:%d", cursor.Row+1, cursor.Col+1)
	statusInfoLen := len(modeStatus + displayFilename + fileStatus + cursorStatus)
	spacesLen := COLS - statusInfoLen
	spaces := strings.Repeat(" ", spacesLen)

	statusBarText := modeStatus + displayFilename + fileStatus + spaces + cursorStatus
	statusBarStyle := tcell.StyleDefault.Background(statusBarColor).Foreground(tcell.ColorBlack)

	drawText(s, ROWS, 0, statusBarStyle, statusBarText)
}

func displayCursor(s tcell.Screen) {
	cursorStyle := tcell.CursorStyle(0)
	s.SetCursorStyle(cursorStyle)
	if len(textBuffer[cursor.Row]) > cursor.Col {
		s.ShowCursor(cursor.Col-offsetCol, cursor.Row-offsetRow)
	} else {
		if len(textBuffer[cursor.Row]) > 0 {
			if mode == normalMode {
				s.ShowCursor(len(textBuffer[cursor.Row])-1, cursor.Row-offsetRow)
			} else {
				s.ShowCursor(len(textBuffer[cursor.Row]), cursor.Row-offsetRow)
			}
		} else {
			s.ShowCursor(0, cursor.Row-offsetRow)
		}
	}
}

func moveCursor(direction string) {
	if direction == "left" {
		if cursor.Col > 0 { cursor.Col-- }
		cursor.ColBuffer = -1
	} else if direction == "right" {
		if cursor.Col < len(textBuffer[cursor.Row])-1 ||
		mode == insertMode && cursor.Col < len(textBuffer[cursor.Row]) {
			cursor.Col++
		}
		cursor.ColBuffer = -1
	} else if direction == "down" {
		if cursor.Row < len(textBuffer)-1 {
			cursor.Row++
			if cursor.ColBuffer < 0 {
				cursor.ColBuffer = cursor.Col
			}
			if cursor.Col > len(textBuffer[cursor.Row])-1 {
				cursor.Col = max(len(textBuffer[cursor.Row])-1, 0)
			} else {
				cursor.Col = cursor.ColBuffer
			}
		}
	} else if direction == "up" {
		if cursor.Row > 0 {
			cursor.Row--
			if cursor.ColBuffer < 0 {
				cursor.ColBuffer = cursor.Col
			}
			if cursor.Col > len(textBuffer[cursor.Row])-1 {
				cursor.Col = max(len(textBuffer[cursor.Row])-1, 0)
			} else {
				cursor.Col = cursor.ColBuffer
			}
		}
	}
}

func loadScreen(s tcell.Screen) {
	COLS, ROWS = s.Size()
	ROWS -= 2 // last two lines are for the status bar
	s.Clear()
	// Set default style
	defStyle := tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault)
	s.SetStyle(defStyle)

	scrollTextBuffer()
	displayTextBuffer(s)
	displayCursor(s)
	displayStatusBar(s)

	s.Show()
}

func insertRune(r rune) {
	rowLen := len(textBuffer[cursor.Row])
	insertRuneRow := make([]rune, rowLen+1)

	if cursor.Col > rowLen { cursor.Col = rowLen }

	insertRuneRow[cursor.Col] = r
	copy(insertRuneRow[:cursor.Col], textBuffer[cursor.Row][:cursor.Col])
	if rowLen > 0 {
		copy(insertRuneRow[cursor.Col+1:], textBuffer[cursor.Row][cursor.Col:])
	}

	textBuffer[cursor.Row] = insertRuneRow
	cursor.Col++
}

func breakLine() {
	newCurrentLine := make([]rune, cursor.Col)
	copy(newCurrentLine, textBuffer[cursor.Row][:cursor.Col])

	newLineLen := len(textBuffer[cursor.Row]) - cursor.Col
	newLine := make([]rune, newLineLen)
	copy(newLine, textBuffer[cursor.Row][cursor.Col:])

	newTextBuffer := make([][]rune, len(textBuffer)+1)
	copy(newTextBuffer[:cursor.Row], textBuffer[:cursor.Row])
	copy(newTextBuffer[cursor.Row+1:], textBuffer[cursor.Row:])

	newTextBuffer[cursor.Row] = newCurrentLine
	newTextBuffer[cursor.Row+1] = newLine

	textBuffer = newTextBuffer

	cursor.Row++
	cursor.Col = 0
}

func removeLineBreak() {
	if cursor.Row == 0 { return }

	aboveRowLen := len(textBuffer[cursor.Row-1])
	currentRowLen := len(textBuffer[cursor.Row])
	newAboveRow := make([]rune, aboveRowLen + currentRowLen)
	copy(newAboveRow[:aboveRowLen], textBuffer[cursor.Row-1])
	copy(newAboveRow[aboveRowLen:], textBuffer[cursor.Row])

	newTextBuffer := make([][]rune, len(textBuffer)-1)
	copy(newTextBuffer[:cursor.Row-1], textBuffer[:cursor.Row-1])
	copy(newTextBuffer[cursor.Row-1:], textBuffer[cursor.Row:])
	newTextBuffer[cursor.Row-1] = newAboveRow

	textBuffer = newTextBuffer

	cursor.Row--
	cursor.Col = aboveRowLen
}

func removeRuneLeft() {
	rowLen := len(textBuffer[cursor.Row])
	if cursor.Col == 0 {
		removeLineBreak()
		return
	}
	removeRuneRow := make([]rune, rowLen-1)

	if cursor.Col > rowLen { cursor.Col = rowLen }

	copy(removeRuneRow[:cursor.Col-1], textBuffer[cursor.Row][:cursor.Col-1])
	copy(removeRuneRow[cursor.Col-1:], textBuffer[cursor.Row][cursor.Col:])

	textBuffer[cursor.Row] = removeRuneRow
	cursor.Col--
}

func changeMode(m string) {
	if m == "normal" {
		mode = normalMode
		if cursor.Col == len(textBuffer[cursor.Row]) {
			cursor.Col--
		}
	} else {
		mode = insertMode
	}

	if m == "append" {
		cursor.Col++
	}
}

// handleEvent realizes actions based on the pressed key and the mode the
// editor is in. It returns true when receives a command to stop the editor.
func handleEvent(s tcell.Screen, ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlC { return true
	} else if ev.Key() == tcell.KeyLeft { moveCursor("left")
	} else if ev.Key() == tcell.KeyDown { moveCursor("down")
	} else if ev.Key() == tcell.KeyUp { moveCursor("up")
	} else if ev.Key() == tcell.KeyRight { moveCursor("right")
	} else if mode == normalMode {
		if ev.Rune() == 'q' || ev.Rune() == 'Q' || ev.Key() == tcell.KeyEscape {
			return true
		} else if ev.Rune() == 'w' { writeFile(sourceFile); modified = false; saved = true
		} else if ev.Rune() == 'h' { moveCursor("left")
		} else if ev.Rune() == 'j' { moveCursor("down")
		} else if ev.Rune() == 'k' { moveCursor("up")
		} else if ev.Rune() == 'l' { moveCursor("right")
		} else if ev.Rune() == 'i' { changeMode("insert")
		} else if ev.Rune() == 'a' { changeMode("append")
		}
	} else if mode == insertMode {
		if ev.Key() == tcell.KeyEsc { changeMode("normal")
		} else {
			if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 { removeRuneLeft()
			} else if ev.Key() == tcell.KeyEnter { breakLine()
			} else { insertRune(ev.Rune()) }
			modified = true
			saved = false
		}
	}

	loadScreen(s)

	return false
}

func runEditor() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	loadScreen(s)

	stop := false
	for !stop {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			loadScreen(s)
		case *tcell.EventKey:
			stop = handleEvent(s, ev)
		}
	}
}
