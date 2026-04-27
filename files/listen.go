package files

func ListenToWrite(fs FileSystem, filepath string, callback func([]byte)) FileSystem {
	return &writeListeningFileSystem{fs, filepath, callback}
}

type writeListeningFileSystem struct {
	FileSystem
	listenFile string
	callback   func([]byte)
}

func (fs *writeListeningFileSystem) Overwrite(loc string, content []byte) error {
	err := fs.FileSystem.Overwrite(loc, content)
	if err != nil {
		return err
	}
	if loc == fs.listenFile {
		fs.callback(content)
	}
	return nil
}
