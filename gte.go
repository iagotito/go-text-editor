package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var ROWS, COLS int
var offsetX, offsetY int

var source_file string
var textBuffer = [][]rune{}

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

func displayTextBuffer(s tcell.Screen) {
	COLS, ROWS = s.Size()
	ROWS -= 2 // last two lines are for the status bar
	var row, col int
	for row = 0; row < ROWS; row++ {
		rowPos := row + offsetY
		for col = 0; col < COLS; col++ {
			colPos := col + offsetX

			if rowPos >= 0 && rowPos < len(textBuffer) && colPos < len(textBuffer[rowPos]) {
				if textBuffer[rowPos][colPos] != '\t' {
					s.SetContent(colPos, rowPos, textBuffer[rowPos][colPos], nil, tcell.StyleDefault)
				} else {
					s.SetContent(colPos, rowPos, ' ', nil, tcell.StyleDefault.Background(tcell.ColorLightGreen))
				}
			} else if rowPos >= len(textBuffer) {
					s.SetContent(0, rowPos, '~', nil, tcell.StyleDefault.Foreground(tcell.ColorBlue))
			}

			if rowPos < len(textBuffer) && colPos == len(textBuffer[rowPos]) {
				s.SetContent(colPos, rowPos, '\n', nil, tcell.StyleDefault)
			}
		}
	}
}

const (
	normalMode = iota
	insertMode
)
var mode int

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
	cursorStatus = "x:y "
	statusInfoLen := len(modeStatus + displayFilename + cursorStatus)
	spacesLen := COLS - statusInfoLen
	spaces := strings.Repeat(" ", spacesLen)

	statusBarText := modeStatus + displayFilename + spaces + cursorStatus
	statusBarStyle := tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tcell.ColorBlack)

	drawText(s, ROWS, 0, statusBarStyle, statusBarText)
}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	row := x
	col := y
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
	}
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

	// Set default style
	defStyle := tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault)
	s.SetStyle(defStyle)

	source_file = os.Args[1]
	readFile(source_file)
	displayTextBuffer(s)
	displayStatusBar(s)

	drawText(s, 18, 0, defStyle, strconv.Itoa(ROWS))
	drawText(s, 1, 0, defStyle, strconv.Itoa(COLS))

	s.Show()

	for {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			displayTextBuffer(s)
			displayStatusBar(s)
		case *tcell.EventKey:
			if ev.Rune() == 'q' || ev.Rune() == 'Q' || ev.Key() == tcell.KeyEscape ||
			ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("No source file provided.")
		return
	}
	runEditor()
}
