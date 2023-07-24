package file

import (
	"io/fs"
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
		Name   string
		From   fs.FileMode
		To     fs.FileMode
		Target ModTarget
		Perm   ModPerm
	}
	cases := []Case{
		{"a", 0000, 0007, ModTargetOther, ModPermAll},
		{"b", 0755, 0757, ModTargetOther, ModPermAll},
		{"c", 0754, 0757, ModTargetOther, ModPermAll},
		{"d", 0750, 0754, ModTargetOther, ModPermRead},
		{"e", 0752, 0752, ModTargetOther, ModPermWrite},
		{"f", 0750, 0751, ModTargetOther, ModPermExec},
		{"g", 0750, 0755, ModTargetOther, ModPermExec | ModPermRead},
		{"h", 0750, 0756, ModTargetOther, ModPermRead | ModPermWrite},
		{"i", 0720, 0766, ModTargetGroup | ModTargetOther, ModPermRead | ModPermWrite},
		{"j", 0120, 0766, ModTargetAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := patch(&tm, "", c.Target, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}

func TestChmodClear(t *testing.T) {
	type Case struct {
		Name   string
		From   fs.FileMode
		To     fs.FileMode
		Target ModTarget
		Perm   ModPerm
	}
	cases := []Case{
		{"a", 0007, 0000, ModTargetOther, ModPermAll},
		{"b", 0755, 0750, ModTargetOther, ModPermAll},
		{"c", 0750, 0750, ModTargetOther, ModPermAll},
		{"d", 0754, 0750, ModTargetOther, ModPermRead},
		{"e", 0756, 0754, ModTargetOther, ModPermWrite},
		{"f", 0757, 0756, ModTargetOther, ModPermExec},
		{"g", 0757, 0752, ModTargetOther, ModPermExec | ModPermRead},
		{"h", 0757, 0751, ModTargetOther, ModPermRead | ModPermWrite},
		{"i", 0725, 0701, ModTargetGroup | ModTargetOther, ModPermRead | ModPermWrite},
		{"j", 0520, 0100, ModTargetAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := clear(&tm, "", c.Target, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}

func TestChmodSet(t *testing.T) {
	type Case struct {
		Name   string
		From   fs.FileMode
		To     fs.FileMode
		Target ModTarget
		Perm   ModPerm
	}
	cases := []Case{
		{"a", 0000, 0700, ModTargetUser, ModPermAll},
		{"b", 0555, 0700, ModTargetUser, ModPermAll},
		{"c", 0421, 0700, ModTargetUser, ModPermAll},
		{"d", 0050, 0400, ModTargetUser, ModPermRead},
		{"e", 0200, 0200, ModTargetUser, ModPermWrite},
		{"f", 0111, 0100, ModTargetUser, ModPermExec},
		{"g", 0111, 0500, ModTargetUser, ModPermExec | ModPermRead},
		{"h", 0210, 0600, ModTargetUser, ModPermRead | ModPermWrite},
		{"i", 0720, 0066, ModTargetGroup | ModTargetOther, ModPermRead | ModPermWrite},
		{"j", 0120, 0666, ModTargetAll, ModPermRead | ModPermWrite},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			tm := testMod{FileMode: c.From}
			err := set(&tm, "", c.Target, c.Perm)
			as.NoError(t, err)
			as.Equal(t, tm.FileMode, c.To)
		})
	}
}
