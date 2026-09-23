package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileMgr(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "filetest")
	fm, err := NewFileMgr(dir, 400)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()

	if !fm.IsNew() {
		t.Error("IsNew() = false, want true")
	}

	blk := newBlockID("filetest", 2)
	p1 := NewPage(fm.BlockSize())

	pos1 := 88
	p1.SetString(88, "abcdefghijklm")

	size := MaxLength(len("abcdefghijklm"))

	pos2 := pos1 + size
	p1.SetInt(pos2, 345)
	if err := fm.Write(blk, p1); err != nil {
		t.Fatal(err)
	}

	p2 := NewPage(fm.BlockSize())
	if err := fm.Read(blk, p2); err != nil {
		t.Fatal(err)
	}
	if got := p2.GetInt(pos2); got != 345 {
		t.Errorf("offset %d contains %d, want 345", pos2, got)
	}
	if got := p2.GetString(pos1); got != "abcdefghijklm" {
		t.Errorf("offset %d contains %q, want %q", pos1, got, "abcdefghijklm")
	}

	// ブロック2に書いたので、0,1,2の3つ存在する
	if n, _ := fm.Length("filetest"); n != 3 {
		t.Errorf("Length = %d, want 3", n)
	}

	blk3, err := fm.Append("filetest")
	if err != nil {
		t.Fatal(err)
	}
	if blk3 != newBlockID("filetest", 3) {
		t.Errorf("Append = %v, want block 3", blk3)
	}
}

func TestNewFileMgrExistingDir(t *testing.T) {
	dir := t.TempDir()
	tmp := filepath.Join(dir, "tmp1")
	if err := os.WriteFile(tmp, []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	fm, err := NewFileMgr(tmp, 400)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()
	if fm.IsNew() {
		t.Error("IsNew() = true, want false")
	}
}

func TestPageNegativeInt(t *testing.T) {
	p := NewPage(16)
	p.SetInt(0, -1)
	if got := p.GetInt(0); got != -1 {
		t.Errorf("GetInt = %d, want -1", got)
	}
}
