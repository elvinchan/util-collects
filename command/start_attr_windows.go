package command

import "syscall"

func detachAttr(attr *syscall.SysProcAttr) {
	attr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}
