// +build !windows

package command

import "syscall"

func detachAttr(attr *syscall.SysProcAttr) {
	attr.Setpgid = true
}
