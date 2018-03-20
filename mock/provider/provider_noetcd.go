// +build !etcd

package provider

import "github.com/rexray/gocsi/middleware/serialvolume"

func newLockProvider() serialvolume.LockProvider {
	return &serialvolume.MemLockProvider{}
}
