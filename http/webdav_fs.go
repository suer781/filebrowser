package fbhttp

import (
	"context"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type webDavFS struct {
	fs afero.Fs
}

func (f *webDavFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return f.fs.MkdirAll(name, perm)
}

func (f *webDavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := f.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &webDavFile{file: file}, nil
}

func (f *webDavFS) RemoveAll(ctx context.Context, name string) error {
	return f.fs.RemoveAll(name)
}

func (f *webDavFS) Rename(ctx context.Context, oldName, newName string) error {
	return f.fs.Rename(oldName, newName)
}

func (f *webDavFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return f.fs.Stat(name)
}

type webDavFile struct {
	file afero.File
}

func (f *webDavFile) Close() error                         { return f.file.Close() }
func (f *webDavFile) Read(p []byte) (n int, err error)      { return f.file.Read(p) }
func (f *webDavFile) Write(p []byte) (n int, err error)     { return f.file.Write(p) }
func (f *webDavFile) Seek(offset int64, whence int) (int64, error) { return f.file.Seek(offset, whence) }
func (f *webDavFile) Readdir(count int) ([]os.FileInfo, error)     { return f.file.Readdir(count) }
func (f *webDavFile) Stat() (os.FileInfo, error)                   { return f.file.Stat() }
