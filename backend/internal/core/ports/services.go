package port 

type DeployService interface {
	AlreadyDeployed(string) (bool, error)
	ValidateURL(string) error
	CloneRepo(string) (string, error)
	BuildRepo(string) (string, error)
	UploadToS3(string) (string, error)
	AddDNSRecord(string) (error)
}

type Store interface {
	DoSomething()
}