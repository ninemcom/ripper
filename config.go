package ripper

type PlayStoreConfiguration struct {
	PackageName                  string
	GCloudServiceAccountFilePath string
}

func (c *PlayStoreConfiguration) ID() string {
	return c.PackageName
}

func (c *PlayStoreConfiguration) Auth() string {
	return c.GCloudServiceAccountFilePath
}
