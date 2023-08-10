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

var source_file string
var textBuffer = [][]rune{}

const (
	normalMode = iota
	insertMode
)
var mode int = 0

type cursorPos struct {
	Row, Col int
}
var cursor = &cursorPos{0, 0}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	row := x
	col := y
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
	}
}

func readFile(filename string) {
	source_file = filename

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
	//var fileStatus string
	var cursorStatus string

	switch mode {
	case normalMode:
		modeStatus = " [ NORMAL ] "
	case insertMode:
		modeStatus = " [ INSERT ] "
	}

	displayFilename = source_file[:8]
	cursorStatus = fmt.Sprintf("%d:%d", cursor.Row, cursor.Col)
	statusInfoLen := len(modeStatus + displayFilename + cursorStatus)
	spacesLen := COLS - statusInfoLen
	spaces := strings.Repeat(" ", spacesLen)

	statusBarText := modeStatus + displayFilename + spaces + cursorStatus
	statusBarStyle := tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tcell.ColorBlack)

	drawText(s, ROWS, 0, statusBarStyle, statusBarText)
}

func displayCursor(s tcell.Screen) {
	cursorStyle := tcell.CursorStyle(0)
	s.SetCursorStyle(cursorStyle)
	if len(textBuffer[cursor.Row]) > cursor.Col {
		s.ShowCursor(cursor.Col-offsetCol, cursor.Row-offsetRow)
	} else {
		if len(textBuffer[cursor.Row]) > 0 {
			s.ShowCursor(len(textBuffer[cursor.Row])-1, cursor.Row-offsetRow)
		} else {
			s.ShowCursor(0, cursor.Row-offsetRow)
		}
	}
}

func moveCursor(direction string) {
	if direction == "left" {
		if cursor.Col > 0 { cursor.Col-- }
	} else if direction == "down" {
		if cursor.Row < len(textBuffer)-1 { cursor.Row++ }
	} else if direction == "up" {
		if cursor.Row > 0 { cursor.Row-- }
	} else if direction == "right" {
		if cursor.Col < len(textBuffer[cursor.Row])-1 { cursor.Col++ }
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

// handleEvent realizes actions based on the pressed key and the mode the
// editor is in. It returns true when receives a command to stop the editor.
func handleEvent(s tcell.Screen, ev *tcell.EventKey) bool {
	if ev.Rune() == 'q' || ev.Rune() == 'Q' || ev.Key() == tcell.KeyEscape ||
	ev.Key() == tcell.KeyCtrlC {
		return true
	} else if ev.Key() == tcell.KeyLeft || mode == normalMode && ev.Rune() == 'h' { moveCursor("left")
	} else if ev.Key() == tcell.KeyDown || mode == normalMode && ev.Rune() == 'j' { moveCursor("down")
	} else if ev.Key() == tcell.KeyUp || mode == normalMode && ev.Rune() == 'k' { moveCursor("up")
	} else if ev.Key() == tcell.KeyRight || mode == normalMode && ev.Rune() == 'l' { moveCursor("right")
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

func main() {
	initLogs()
	if len(os.Args) == 1 {
		log.Println("No source file provided.")
		fmt.Println("No source file provided.")
		return
	}
	source_file = os.Args[1]

	readFile(source_file)
	runEditor()
}

func initLogs() {
 logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	log.SetOutput(logFile)
}
