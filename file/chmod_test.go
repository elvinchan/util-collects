package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/elvinchan/util-collects/as"
)

type testMod struct {
	FileMode fs.FileMode
}

func (t testMod) Stat(name string) (fs.FileMode, error) {
	return t.FileMode, nil
}

func (t *testMod) Chmod(name string, mode fs.FileMode) error {
	t.FileMode = mode
	return nil
}

func TestChmodPatch(t *testing.T) {
	type Case struct {
		Name string
		From fs.FileMode
		To   fs.FileMode
		Role ModRole
		Perm ModPerm
	}
	cases := []Case{
		{"a", 0000, 0007, ModRoleOther, ModPermAll},
		{"b", 0755, 0757, ModRoleOther, ModPermAll},
		{"c", 0754, 0757, ModRoleOther, ModPermAll},
		{"d", 0750, 0754, ModRoleOther, ModPermRead},
		{"e", 0752, 0752, ModRoleOther, ModPermWrite},
		{"f", 0750, 0751, ModRoleOther, ModPermExec},
		{"g", 0750, 0755, ModRoleOther, ModPermExec | ModPermRead},
		{"h", 0750, 0756, ModRoleOther, ModPermRead | ModPermWrite},
		{"i", 0720, 0766, ModRoleGroup | ModRoleOther, ModPermRead | ModPermWrite},
		{"j", 0120, 0766, ModRoleAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := modPatch(&tm, "", c.Role, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}

func TestChmodClear(t *testing.T) {
	type Case struct {
		Name string
		From fs.FileMode
		To   fs.FileMode
		Role ModRole
		Perm ModPerm
	}
	cases := []Case{
		{"a", 0007, 0000, ModRoleOther, ModPermAll},
		{"b", 0755, 0750, ModRoleOther, ModPermAll},
		{"c", 0750, 0750, ModRoleOther, ModPermAll},
		{"d", 0754, 0750, ModRoleOther, ModPermRead},
		{"e", 0756, 0754, ModRoleOther, ModPermWrite},
		{"f", 0757, 0756, ModRoleOther, ModPermExec},
		{"g", 0757, 0752, ModRoleOther, ModPermExec | ModPermRead},
		{"h", 0757, 0751, ModRoleOther, ModPermRead | ModPermWrite},
		{"i", 0725, 0701, ModRoleGroup | ModRoleOther, ModPermRead | ModPermWrite},
		{"j", 0520, 0100, ModRoleAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := modClear(&tm, "", c.Role, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}

func TestChmodSet(t *testing.T) {
	type Case struct {
		Name string
		From fs.FileMode
		To   fs.FileMode
		Role ModRole
		Perm ModPerm
	}
	cases := []Case{
		{"a", 0000, 0700, ModRoleUser, ModPermAll},
		{"b", 0555, 0700, ModRoleUser, ModPermAll},
		{"c", 0421, 0700, ModRoleUser, ModPermAll},
		{"d", 0050, 0400, ModRoleUser, ModPermRead},
		{"e", 0200, 0200, ModRoleUser, ModPermWrite},
		{"f", 0111, 0100, ModRoleUser, ModPermExec},
		{"g", 0111, 0500, ModRoleUser, ModPermExec | ModPermRead},
		{"h", 0210, 0600, ModRoleUser, ModPermRead | ModPermWrite},
		{"i", 0720, 0066, ModRoleGroup | ModRoleOther, ModPermRead | ModPermWrite},
		{"j", 0120, 0666, ModRoleAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := modSet(&tm, "", c.Role, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}

func TestModPatchWalk(t *testing.T) {
	u := syscall.Umask(0)
	defer syscall.Umask(u)

	root := t.TempDir()
	dir := filepath.Join(root, "mod/patch/walk")
	err := os.MkdirAll(dir, 0755)
	as.NoError(t, err)

	f := filepath.Join(dir, "testfile")
	_, err = os.Create(f)
	as.NoError(t, err)

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		as.NoError(t, err)
		// user must have read and exec permission
		// or you cannot walk it again
		err = os.Chmod(path, 0511)
		as.NoError(t, err)
		return nil
	})
	as.NoError(t, err)

	err = os.Chmod(f, 0400)
	as.NoError(t, err)

	err = ModPatchWalk(root, ModRoleUser|ModRoleOther, ModPermWrite, ModTargetAll)
	as.NoError(t, err)

	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		as.NoError(t, err)
		if info.IsDir() {
			as.Equal(t, fs.FileMode(0713).Perm(), info.Mode().Perm())
		} else {
			as.Equal(t, fs.FileMode(0602).Perm(), info.Mode().Perm())
		}
		return nil
	})
	as.NoError(t, err)
}
