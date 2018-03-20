// +build etcd

package provider

import (
	"github.com/rexray/gocsi/middleware/serialvolume"
	"github.com/rexray/gocsi/middleware/serialvolume/etcd"
)

func newLockProvider() serialvolume.LockProvider {
	return &etcd.LockProvider{}
}
