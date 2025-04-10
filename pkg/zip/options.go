package zip

type ZipOption struct {
	ProgressBar    bool
	TotalFileCount int
}

var DefaultZipOption = ZipOption{
	ProgressBar: false,
}
