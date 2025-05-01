package service

type File interface {
	FileService() FileService
}

type file struct{}

func NewFile() File {
	return &file{}
}

func (ps *file) FileService() FileService {
	return &fileService{}
}
