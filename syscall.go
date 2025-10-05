package minivm

type Syscall = int

const (
	SYS_EXIT Syscall = iota
	SYS_WRITE
	SYS_READ
)
