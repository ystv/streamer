package gomux

import (
	"fmt"
	"io"
	"strings"
)

type Pane struct {
	Number   int
	commands []string
	Window   *Window
}

func NewPane(number int, window *Window) *Pane {
	p := new(Pane)
	p.Number = number
	p.commands = make([]string, 0)
	p.Window = window
	return p
}

type SplitAttr struct {
	Directory string
}

func (this *Pane) Exec(command string) {
	fmt.Fprintf(this.Window.Session.Writer, "tmux send-keys -t \"%s\" \"%s\" %s\n", this.getTargetName(), strings.Replace(command, "\"", "\\\"", -1), "C-m")
}

func (this *Pane) Vsplit() *Pane {
	fmt.Fprint(this.Window.Session.Writer, splitWindow{h: true, t: this.getTargetName()})
	return this.Window.AddPane(this.Number + 1)
}

func (this *Pane) VsplitWAttr(attr SplitAttr) *Pane {
	var c string
	if attr.Directory != "" {
		c = attr.Directory
	} else if this.Window.Directory != "" {
		c = this.Window.Directory
	} else if this.Window.Session.Directory != "" {
		c = this.Window.Session.Directory
	}

	fmt.Fprint(this.Window.Session.Writer, splitWindow{h: true, t: this.getTargetName(), c: c})
	return this.Window.AddPane(this.Number + 1)
}

func (this *Pane) Split() *Pane {
	fmt.Fprint(this.Window.Session.Writer, splitWindow{v: true, t: this.getTargetName()})
	return this.Window.AddPane(this.Number + 1)
}

func (this *Pane) SplitWAttr(attr SplitAttr) *Pane {
	var c string
	if attr.Directory != "" {
		c = attr.Directory
	} else if this.Window.Directory != "" {
		c = this.Window.Directory
	} else if this.Window.Session.Directory != "" {
		c = this.Window.Session.Directory
	}

	fmt.Fprint(this.Window.Session.Writer, splitWindow{v: true, t: this.getTargetName(), c: c})
	return this.Window.AddPane(this.Number + 1)
}

func (this *Pane) ResizeRight(num int) {
	this.resize("R", num)
}

func (this *Pane) ResizeLeft(num int) {
	this.resize("L", num)
}

func (this *Pane) ResizeUp(num int) {
	this.resize("U", num)
}

func (this *Pane) ResizeDown(num int) {
	this.resize("U", num)
}

func (this *Pane) resize(prefix string, num int) {
	fmt.Fprintf(this.Window.Session.Writer, "tmux resize-pane -t \"%s\" -%s %v\n", this.getTargetName(), prefix, fmt.Sprint(num))
}

func (this *Pane) getTargetName() string {
	return this.Window.Session.Name + ":" + fmt.Sprint(this.Window.Number) + "." + fmt.Sprint(this.Number)
}

// Window Represent a tmux Window. You usually should not create an instance of Window directly.
type Window struct {
	Number         int
	Name           string
	Directory      string
	Session        *Session
	panes          []*Pane
	split_commands []string
}

type WindowAttr struct {
	Name      string
	Directory string
}

func createWindow(number int, attr WindowAttr, session *Session) *Window {
	w := new(Window)
	w.Name = attr.Name
	w.Directory = attr.Directory
	w.Number = number
	w.Session = session
	w.panes = make([]*Pane, 0)
	w.split_commands = make([]string, 0)
	w.AddPane(0)

	if number != 0 {
		fmt.Fprint(session.Writer, newWindow{t: w.t(), n: w.Name, c: attr.Directory})
	}

	fmt.Fprint(session.Writer, renameWindow{t: w.t(), n: w.Name})
	return w
}

func (this *Window) t() string {
	return fmt.Sprintf("-t \"%s:%s\"", this.Session.Name, fmt.Sprint(this.Number))
}

// Create a new Pane and add to this Window
func (this *Window) AddPane(withNumber int) *Pane {
	pane := NewPane(withNumber, this)
	this.panes = append(this.panes, pane)
	return pane
}

// Find and return the Pane object by its index in the panes slice
func (this *Window) Pane(number int) *Pane {
	return this.panes[number]
}

// Executes a command on the first pane of this Window
//
// // example
// // example
func (this *Window) Exec(command string) {
	this.Pane(0).Exec(command)
}

func (this *Window) Select() {
	fmt.Fprint(this.Session.Writer, selectWindow{t: this.Session.Name + ":" + fmt.Sprint(this.Number)})
}

// Session represents a tmux Session.
//
// Use the method NewSession to create a Session instance.
type Session struct {
	Name               string
	Directory          string
	windows            []*Window
	directory          string
	next_window_number int
	Writer             io.Writer
}

// Creates a new Tmux Session. It will kill any existing Session with the provided name.
func NewSession(name string, writer io.Writer) *Session {
	p := SessionAttr{
		Name: name,
	}
	return NewSessionAttr(p, writer)
}

type SessionAttr struct {
	Name      string
	Directory string
}

// Creates a new Tmux Session based on NewSessionAttr. It will kill any existing Session with the provided name.
func NewSessionAttr(p SessionAttr, writer io.Writer) *Session {
	s := new(Session)
	s.Writer = writer
	s.Name = p.Name
	s.Directory = p.Directory
	s.windows = make([]*Window, 0)

	fmt.Fprint(writer, newSession{d: true, s: p.Name, c: p.Directory, n: "tmp"})
	return s
}

// KillSession sends a command to kill the tmux Session
func KillSession(name string, writer io.Writer) {
	fmt.Fprint(writer, killSession{t: name})
}

// Creates Window with provided name for this Session
func (this *Session) AddWindow(name string) *Window {

	attr := WindowAttr{
		Name: name,
	}

	return this.AddWindowAttr(attr)
}

// Creates Window with provided name for this Session
func (this *Session) AddWindowAttr(attr WindowAttr) *Window {
	w := createWindow(this.next_window_number, attr, this)
	this.windows = append(this.windows, w)
	this.next_window_number = this.next_window_number + 1
	return w
}
