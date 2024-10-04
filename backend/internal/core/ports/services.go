package port 

type DeployService interface {
	Deploy() 
	ValidateURL(string) error
	CloneRepo(string) (string, error)
	BuildRepo(string) (string, error)
	UploadToS3(string) (string, error)
	AddDNSRecord(string) (error)
}

type PingService interface {
	SayHello()
}