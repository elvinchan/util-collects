package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

type ModTarget uint32

const (
	ModTargetUser  = syscall.S_IRWXU // 0b111000000
	ModTargetGroup = syscall.S_IRWXG // 0b000111000
	ModTargetOther = syscall.S_IRWXO // 0b000000111
	ModTargetAll   = ModTargetUser | ModTargetGroup | ModTargetOther
)

type ModPerm uint32

const (
	ModPermRead  = syscall.S_IRUSR | syscall.S_IRGRP | syscall.S_IROTH // 0b100100100
	ModPermWrite = syscall.S_IWUSR | syscall.S_IWGRP | syscall.S_IWOTH // 0b010010010
	ModPermExec  = syscall.S_IXUSR | syscall.S_IXGRP | syscall.S_IXOTH // 0b001001001
	ModPermAll   = ModPermRead | ModPermWrite | ModPermExec
)

func ModPatch(path string, t ModTarget, p ModPerm) error {
	return modPatch(defaultMod{}, path, t, p)
}

func ModClear(path string, t ModTarget, p ModPerm) error {
	return modClear(defaultMod{}, path, t, p)
}

func ModSet(path string, t ModTarget, p ModPerm) error {
	return modSet(defaultMod{}, path, t, p)
}

func ModPatchWalk(path string, t ModTarget, p ModPerm) error {
	return modWalkPatch(defaultMod{}, path, t, p)
}

func ModPatchClear(path string, t ModTarget, p ModPerm) error {
	return modWalkClear(defaultMod{}, path, t, p)
}

func ModPatchSet(path string, t ModTarget, p ModPerm) error {
	return modWalkSet(defaultMod{}, path, t, p)
}

type ModProvider interface {
	Stat(name string) (fs.FileMode, error)
	Chmod(name string, mode fs.FileMode) error
}

type defaultMod struct{}

func (defaultMod) Stat(name string) (fs.FileMode, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return fs.FileMode(0), err
	}
	return fi.Mode(), err
}

func (defaultMod) Chmod(name string, mode fs.FileMode) error {
	return os.Chmod(name, mode)
}

func modPatch(m ModProvider, path string, t ModTarget, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := mod | (fs.FileMode(t) & fs.FileMode(p))
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func modClear(m ModProvider, path string, t ModTarget, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := mod &^ (fs.FileMode(t) & fs.FileMode(p))
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func modSet(m ModProvider, path string, t ModTarget, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := fs.FileMode(t) & fs.FileMode(p)
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func modWalkPatch(m ModProvider, path string, t ModTarget, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			p |= ModPermExec
		}
		mod := info.Mode()
		target := mod | (fs.FileMode(t) & fs.FileMode(p))
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}

func modWalkClear(m ModProvider, path string, t ModTarget, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			p &^= ModPermExec
		}
		mod := info.Mode()
		target := mod &^ (fs.FileMode(t) & fs.FileMode(p))
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}

func modWalkSet(m ModProvider, path string, t ModTarget, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			p |= ModPermExec
		}
		mod := info.Mode()
		target := fs.FileMode(t) & fs.FileMode(p)
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}
