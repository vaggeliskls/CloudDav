package storage

import (
	"golang.org/x/net/webdav"
)

// NewLocal returns a webdav.FileSystem backed by the local filesystem at root.
// It uses the built-in webdav.Dir implementation.
func NewLocal(root string) webdav.FileSystem {
	return webdav.Dir(root)
}
