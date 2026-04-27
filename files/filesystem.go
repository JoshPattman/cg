package files

type FileSystem interface {
	ListDir(loc string) ([]string, error)
	Overwrite(loc string, content []byte) error
	Read(loc string) ([]byte, error)
	Delete(loc string) error
}
