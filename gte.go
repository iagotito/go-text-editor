package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

var ROWS, COLS int
var offsetX, offsetY int

var textBuffer = [][]rune{
	[]rune("Hello more text to text hahaha"),
	[]rune("World"),
}

func displayTextBuffer(s tcell.Screen) {
	COLS, ROWS = s.Size()
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

	//drawText(s, 6, 20, defStyle, "Go Text Editor")
	//drawText(s, 10, 20, defStyle, "Press 'q' to exit")

	displayTextBuffer(s)

	s.Show()

	for {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
			displayTextBuffer(s)
		case *tcell.EventKey:
			if ev.Rune() == 'q' || ev.Rune() == 'Q' || ev.Key() == tcell.KeyEscape ||
			ev.Key() == tcell.KeyCtrlC {
				return
			}
		}
	}
}

func main() {
	runEditor()
}
