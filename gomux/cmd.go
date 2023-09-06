package gomux

import "fmt"

type killSession struct {
	t string
}

func (this killSession) String() string {
	return fmt.Sprintf("tmux kill-Session -t \"%s\" > /dev/null 2>&1\n", this.t)
}

type newSession struct {
	d bool
	s string
	n string
	c string
}

func (this newSession) String() string {
	cmd := "tmux new-Session"
	if this.d == true {
		cmd += " -d"
	}
	if this.s != "" {
		cmd += " -s \"" + this.s + "\""
	}
	if this.n != "" {
		cmd += " -n " + this.n
	}
	if this.c != "" {
		cmd += " -c " + this.c
	}
	return cmd + "\n"
}

type splitWindow struct {
	h bool
	v bool
	t string
	c string
}

func (this splitWindow) String() string {
	cmd := "tmux split-Window"
	if this.h == true {
		cmd += " -h"
	}
	if this.v == true {
		cmd += " -v"
	}
	if this.t != "" {
		cmd += " -t \"" + this.t + "\""
	}
	if this.c != "" {
		cmd += " -c " + this.c
	}
	return cmd + "\n"
}

type newWindow struct {
	t string
	n string
	c string
}

func (this newWindow) String() string {
	cmd := "tmux new-Window"
	if this.t != "" {
		cmd += " " + this.t
	}
	if this.n != "" {
		cmd += " -n \"" + this.n + "\""
	}

	if this.c != "" {
		cmd += " -c " + this.c
	}
	return cmd + "\n"
}

type renameWindow struct {
	t string
	n string
}

func (this renameWindow) String() string {
	cmd := "tmux rename-Window"
	if this.t != "" {
		cmd += " " + this.t
	}
	if this.n != "" {
		cmd += " \"" + this.n + "\""
	}

	return cmd + "\n"
}

type selectWindow struct {
	t string
}

func (this selectWindow) String() string {
	cmd := "tmux select-Window"
	if this.t != "" {
		cmd += " -t \"" + this.t + "\""
	}

	return cmd + "\n"
}
