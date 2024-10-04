package port 

type DeployService interface {
	Deploy() 
	// ValidateURL()
	// BuildProject()
	// UploadToS3()
	// MapDNS()
}

type PingService interface {
	SayHello()
}