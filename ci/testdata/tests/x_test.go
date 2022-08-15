package tests

import (
	"os"
	"syscall"
	"testing"
)

func TestX(t *testing.T) {
	t.Log("hallo")
	info, err := os.Stat("../hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatal("hello.txt is a directory")
	}
	stat := info.Sys().(*syscall.Stat_t)
	if int(stat.Uid) != os.Getuid() {
		t.Errorf("hello.txt has owner %d, expected %d", stat.Uid, os.Getuid())
	}
	if int(stat.Gid) != os.Getgid() {
		t.Errorf("hello.txt has group %d, expected %d", stat.Gid, os.Getgid())
	}
}
