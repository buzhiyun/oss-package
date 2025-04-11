package zip

type ZipOption struct {
	ProgressBar    bool
	TotalFileCount int
	ChannelBar     bool
}

var DefaultZipOption = ZipOption{
	ProgressBar: false,
	ChannelBar:  false,
}
