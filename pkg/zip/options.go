package zip

type ZipOption struct {
	ProgressBar    bool
	TotalFileCount int
	ChannelBar     bool
	ZipLevel       int
}

var DefaultZipOption = ZipOption{
	ProgressBar: false,
	ChannelBar:  false,
	ZipLevel:    1,
}
