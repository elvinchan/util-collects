package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

type ModRole uint32

const (
	ModRoleUser  = syscall.S_IRWXU // 0b111000000
	ModRoleGroup = syscall.S_IRWXG // 0b000111000
	ModRoleOther = syscall.S_IRWXO // 0b000000111
	ModRoleAll   = ModRoleUser | ModRoleGroup | ModRoleOther
)

type ModPerm uint32

const (
	ModPermRead  = syscall.S_IRUSR | syscall.S_IRGRP | syscall.S_IROTH // 0b100100100
	ModPermWrite = syscall.S_IWUSR | syscall.S_IWGRP | syscall.S_IWOTH // 0b010010010
	ModPermExec  = syscall.S_IXUSR | syscall.S_IXGRP | syscall.S_IXOTH // 0b001001001
	ModPermAll   = ModPermRead | ModPermWrite | ModPermExec
)

type ModTarget uint32

const (
	ModTargetFile = 1
	ModTargetDir  = ModTargetFile << 1
	ModTargetAll  = ModTargetFile | ModTargetDir
)

type Mod struct {
	ModProvider
	target ModTarget
}

func ModPatch(path string, r ModRole, p ModPerm) error {
	return modPatch(defaultModProvider, path, r, p)
}

func ModClear(path string, r ModRole, p ModPerm) error {
	return modClear(defaultModProvider, path, r, p)
}

func ModSet(path string, r ModRole, p ModPerm) error {
	return modSet(defaultModProvider, path, r, p)
}

func ModPatchWalk(path string, tr ModRole, p ModPerm, t ModTarget) error {
	m := &Mod{
		ModProvider: defaultModProvider,
		target:      t,
	}
	return m.patchWalk(path, tr, p)
}

func ModClearWalk(path string, tr ModRole, p ModPerm, t ModTarget) error {
	m := &Mod{
		ModProvider: defaultModProvider,
		target:      t,
	}
	return m.clearWalk(path, tr, p)
}

func ModSetWalk(path string, tr ModRole, p ModPerm, t ModTarget) error {
	m := &Mod{
		ModProvider: defaultModProvider,
		target:      ModTargetAll,
	}
	return m.setWalk(path, tr, p)
}

type ModProvider interface {
	Stat(name string) (fs.FileMode, error)
	Chmod(name string, mode fs.FileMode) error
}

type modProvider struct{}

func (modProvider) Stat(name string) (fs.FileMode, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return fs.FileMode(0), err
	}
	return fi.Mode(), err
}

func (modProvider) Chmod(name string, mode fs.FileMode) error {
	return os.Chmod(name, mode)
}

var defaultModProvider = modProvider{}

func modPatch(m ModProvider, path string, r ModRole, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := mod | (fs.FileMode(r) & fs.FileMode(p))
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func modClear(m ModProvider, path string, r ModRole, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := mod &^ (fs.FileMode(r) & fs.FileMode(p))
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func modSet(m ModProvider, path string, r ModRole, p ModPerm) error {
	mod, err := m.Stat(path)
	if err != nil {
		return err
	}
	target := fs.FileMode(r) & fs.FileMode(p)
	if target == mod { // already satisfy
		return nil
	}
	return m.Chmod(path, target)
}

func (m *Mod) patchWalk(path string, r ModRole, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && (m.target&ModTargetFile) == 0 {
			return nil
		} else if info.IsDir() && (m.target&ModTargetDir) == 0 {
			return nil
		}
		mod := info.Mode()
		target := mod | (fs.FileMode(r) & fs.FileMode(p))
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}

func (m *Mod) clearWalk(path string, r ModRole, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && (m.target&ModTargetFile) == 0 {
			return nil
		} else if info.IsDir() && (m.target&ModTargetDir) == 0 {
			return nil
		}
		mod := info.Mode()
		target := mod &^ (fs.FileMode(r) & fs.FileMode(p))
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}

func (m *Mod) setWalk(path string, r ModRole, p ModPerm) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && (m.target&ModTargetFile) == 0 {
			return nil
		} else if info.IsDir() && (m.target&ModTargetDir) == 0 {
			return nil
		}
		mod := info.Mode()
		target := fs.FileMode(r) & fs.FileMode(p)
		if target == mod { // already satisfy
			return nil
		}
		return m.Chmod(path, target)
	})
}
