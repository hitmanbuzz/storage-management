package storage

type User struct {
	Name  string `json:"username"`
	Token string `json:"-"`
}

type File struct {
	Name   string `json:"filename"`
	Size   int64  `json:"filesize"`
	Ext    string `json:"-"`
	IsErr  bool   `json:"-"`
	Status bool   `json:"status"`
}

func NewFile(name string, ext string) *File {
	return &File{
		Name:   name,
		Size:   0,
		Ext:    ext,
		IsErr:  false,
		Status: true,
	}
}

type Upload struct {
	User     User    `json:"user"`
	Files    []*File `json:"files"`
	CurrFile *File   `json:"-"`
}

func NewUpload() *Upload {
	return &Upload{
		Files: make([]*File, 0),
	}
}

func (u *Upload) SetUsername(name string) {
	u.User.Name = name
}

func (u *Upload) SetToken(token string) {
	u.User.Token = token
}

func (u *Upload) AddFile(file *File) {
	u.Files = append(u.Files, file)
}

func (u *Upload) UpdateCurrFile(file *File) {
	if u.CurrFile == nil {
		u.CurrFile = file
		return
	}

	if u.CurrFile.Name != file.Name {
		u.AddFile(u.CurrFile)
		u.CurrFile = file
		return
	}
}
