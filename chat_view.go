package chat_sample_view

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func NewChatView(prompt string) *ChatView {
	chat := &ChatView{
		prompt:      prompt,
		inputChan:   make(chan rune),
		removeChan:  make(chan struct{}),
		messageChan: make(chan fmt.Stringer),
	}
	go chat.render()
	return chat
}

type action func(text string)

type ChatView struct {
	prompt      string
	inputs      []rune
	messages    []fmt.Stringer
	EnterAction action
	inputChan   chan rune
	removeChan  chan struct{}
	messageChan chan fmt.Stringer
}

func (c ChatView) Init() error {
	return termbox.Init()
}

func (c ChatView) Clear() error {
	return termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func (c ChatView) Flush() {
	termbox.Flush()
}

func (c ChatView) Close() {
	termbox.Close()
}

func (c *ChatView) PollInput() {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return
			case termbox.KeyEnter:
				s := string(c.inputs)
				c.inputs = []rune{}
				c.EnterAction(s)
			case termbox.KeyBackspace, termbox.KeyBackspace2, termbox.KeyDelete:
				c.removeChan <- struct{}{}
			case termbox.KeyTab:
				c.inputChan <- '\t'
			case termbox.KeySpace:
				c.inputChan <- ' '
			default:
				c.inputChan <- ev.Ch
			}
		}
	}
}

func (c ChatView) Draw() {
	c.Clear()
	c.drawInput()
	for i := len(c.messages); i > 0; i-- {
		c.drawLine(len(c.messages)-i+1, c.messages[i-1].String())
	}
	c.Flush()
}

func (c *ChatView) Message(message fmt.Stringer) {
	c.messageChan <- message
}

func (c *ChatView) render() {
	for {
		select {
		case input := <-c.inputChan:
			c.inputs = append(c.inputs, input)
		case <-c.removeChan:
			c.inputs = c.inputs[:len(c.inputs)-1]
		case message := <-c.messageChan:
			c.messages = append(c.messages, message)
		}
		c.Draw()
	}

}

func (c ChatView) drawInput() {
	input := fmt.Sprintf("%s%s", c.prompt, string(c.inputs))
	c.drawLine(0, input)
	termbox.SetCursor(runewidth.StringWidth(input), 0)
}

func (c ChatView) drawLine(line int, text string) {
	var x int
	for _, c := range text {
		termbox.SetCell(x, line, c, termbox.ColorDefault, termbox.ColorDefault)
		x = x + runewidth.RuneWidth(c)
	}
}
