package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

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

	drawText(s, 6, 20, defStyle, "Go Text Editor")
	drawText(s, 10, 20, defStyle, "Press 'q' to exit")

	s.Show()

	for {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
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
