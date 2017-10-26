package mount_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/thecodeteam/gocsi/mount"
)

func TestBindMount(t *testing.T) {
	src, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	tgt, err := ioutil.TempDir("", "")
	if err != nil {
		os.RemoveAll(src)
		t.Fatal(err)
	}
	if err := mount.EvalSymlinks(&src); err != nil {
		os.RemoveAll(tgt)
		os.RemoveAll(src)
		t.Fatal(err)
	}
	if err := mount.EvalSymlinks(&tgt); err != nil {
		os.RemoveAll(tgt)
		os.RemoveAll(src)
		t.Fatal(err)
	}
	defer func() {
		mount.Unmount(tgt)
		os.RemoveAll(tgt)
		os.RemoveAll(src)
	}()
	if err := mount.BindMount(src, tgt); err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Logf("bind mount success: source=%s, target=%s", src, tgt)
	mounts, err := mount.GetMounts(context.TODO(), nil)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	success := false
	for _, m := range mounts {
		if m.Source == src && m.Path == tgt {
			success = true
		}
		t.Logf("%+v", m)
	}
	if !success {
		t.Errorf("unable to find bind mount: src=%s, tgt=%s", src, tgt)
		t.Fail()
	}
}

func TestGetMounts(t *testing.T) {
	mounts, err := mount.GetMounts(context.TODO(), nil)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	for _, m := range mounts {
		t.Logf("%+v", m)
	}
}
