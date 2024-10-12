package services

import (
	"bytes"
	"codeflare/internal/config"
	"codeflare/internal/core/domain"
	"codeflare/internal/core/ports"
	"codeflare/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type deployService struct {
	db               ports.Repository
	buildQ           chan uint
	deployQ          chan uint
	deployedProjects map[string]time.Time
}

func NewDeployService(db ports.Repository, chsize int) ports.DeployService {
	return &deployService{
		db:      db,
		buildQ:  make(chan uint, chsize),
		deployQ: make(chan uint, chsize),
	}
}

func (s *deployService) AlreadyDeployed(url string) (bool, error) {
	val, err := s.db.FindRepo(url)
	if err != nil {
		return false, err
	}
	return val, nil
}

// Add this method to queue builds
func (s *deployService) QueueBuild(projectID uint) {
	s.buildQ <- projectID
}

// Modify BuildRepo to push to deployQ after successful build
func (s *deployService) BuildRepo() {
	for {
		projectId := <-s.buildQ
		proj, err := s.db.GetProject(projectId)

		if err != nil {
			fmt.Printf("Error getting project: %v\n", err)
			continue
		}
		fmt.Println("strarted building:", projectId)

		dir := "./projects/" + proj.Name + proj.ProjectDirectory

		if err := s.buildProject(dir); err != nil {
			fmt.Printf("Error building project: %v\n", err)
			s.db.UpdateStatus(projectId, domain.Failed)
			continue
		}

		s.db.UpdateStatus(projectId, domain.Building)
		fmt.Println("project qd for deployment:", projectId)
		s.deployQ <- projectId
	}
}

// Add this method to build the project
func (s *deployService) buildProject(dir string) error {
	// Install dependencies
	fmt.Println("npm i")
	installCmd := exec.Command("npm", "install")
	installCmd.Dir = dir
	if output, err := installCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s, error: %v", string(output), err)
	}

	// Build the project
	buildCmd := exec.Command("npm", "run", "build")
	fmt.Println("npm build")
	buildCmd.Dir = dir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build project: %s, error: %v", string(output), err)
	}

	return nil
}

// Deploy
func (s *deployService) Deploy() {
	for {
		projectID := <-s.deployQ
		project, err := s.db.GetProject(projectID)
		if err != nil {
			fmt.Printf("Error getting project: %v\n", err)
			continue
		}
		fmt.Println("deploying proj:", projectID)

		s.db.UpdateStatus(projectID, domain.Deploying)
		err = s.deployProject(project)
		if err != nil {
			fmt.Printf("Error deploying project: %v\n", err)
			s.db.UpdateStatus(projectID, domain.Failed)
		} else {
			s.db.UpdateStatus(projectID, domain.Deployed)
			s.deployedProjects[project.Name] = time.Now()
			fmt.Println("Project deployed:", project.Name)

			// Clean up local files after deploy
			if err := s.cleanupLocalFiles(project); err != nil {
				fmt.Printf("Error cleaning up local files: %v\n", err)
			}
		}
	}
}

// func to clean up local files
func (s *deployService) cleanupLocalFiles(project *domain.Project) error {
	dir := project.ProjectDirectory
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove project directory: %v", err)
	}
	return nil
}

func (s *deployService) StartCleanupTicker() {
	ticker := time.NewTicker(30 * time.Minute) // Check every 30
	go func() {
		for {
			<-ticker.C
			s.cleanupOldDeployments()
		}
	}()
}

func (s *deployService) cleanupOldDeployments() {
	now := time.Now()
	for projectName, deployTime := range s.deployedProjects {
		if now.Sub(deployTime) > 1*time.Hour {
			// Delete the project
			project, err := s.db.GetProjectByName(projectName)
			if err != nil {
				fmt.Printf("Error getting project %s: %v\n", projectName, err)
				continue
			}

			err = s.DeleteProject(project.ID)
			if err != nil {
				fmt.Printf("Error deleting project %s: %v\n", projectName, err)
			} else {
				fmt.Printf("Project %s deleted successfully\n", projectName)
				delete(s.deployedProjects, projectName) // Remove from the map
			}
		}
	}
}

func (s *deployService) deployProject(project *domain.Project) error {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("ap-south-1"))
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	fmt.Println("config loaded")

	// Create S3 service client
	svc := s3.NewFromConfig(cfg)
	bucket := project.Name + ".nymbus.xyz"

	err = s.ensureBucketExists(svc, bucket)
	if err != nil {
		return err
	}
	fmt.Println("bucket there")
	buildDir := "./projects/" + project.Name + project.ProjectDirectory
	staticSiteURL, err := s.uploadFiles(svc, buildDir, bucket)
	if err != nil {
		return err
	}

	err = s.db.UpdateURL(project.ID, staticSiteURL)
	if err != nil {
		return err
	}

	return s.AddDNSRecord(staticSiteURL, project.Name)
}

func (s *deployService) loadAWSConfig() (aws.Config, error) {
	return awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("ap-south-1"))
}

func (s *deployService) ensureBucketExists(svc *s3.Client, bucket string) error {
	_, err := svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		return nil // Bucket already exists
	}
	fmt.Println("bucket not there")
	err = s.createBucket(svc, bucket)
	if err != nil {
		return err
	}
	fmt.Println("created")

	err = s.disablePublicAccessBlock(svc, bucket)
	if err != nil {
		return err
	}
	fmt.Println("pub access")

	err = s.enableStaticSiteHosting(svc, bucket)
	if err != nil {
		return err
	}
	fmt.Println("enable site hosting")

	return s.setPublicAccessPolicy(svc, bucket)
}

func (s *deployService) createBucket(svc *s3.Client, bucket string) error {
	_, err := svc.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: "ap-south-1",
		},
	})
	return err
}

func (s *deployService) disablePublicAccessBlock(svc *s3.Client, bucket string) error {
	_, err := svc.PutPublicAccessBlock(context.TODO(), &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(false),
			IgnorePublicAcls:      aws.Bool(false),
			BlockPublicPolicy:     aws.Bool(false),
			RestrictPublicBuckets: aws.Bool(false),
		},
	})
	return err
}

func (s *deployService) enableStaticSiteHosting(svc *s3.Client, bucket string) error {
	_, err := svc.PutBucketWebsite(context.TODO(), &s3.PutBucketWebsiteInput{
		Bucket: aws.String(bucket),
		WebsiteConfiguration: &s3Types.WebsiteConfiguration{
			IndexDocument: &s3Types.IndexDocument{
				Suffix: aws.String("index.html"),
			},
			ErrorDocument: &s3Types.ErrorDocument{
				Key: aws.String("index.html"),
			},
		},
	})
	return err
}

func (s *deployService) setPublicAccessPolicy(svc *s3.Client, bucket string) error {
	publicPolicy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicReadGetObject",
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			}
		]
	}`, bucket)

	_, err := svc.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(publicPolicy),
	})
	return err
}

func (s *deployService) uploadFiles(svc *s3.Client, dir, bucket string) (string, error) {
	files, err := utils.GetFilePaths(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get file paths: %v", err)
	}

	for _, path := range files {
		err := s.uploadFile(svc, path, bucket)
		if err != nil {
			return "", err
		}
	}

	staticSiteURL := fmt.Sprintf("http://%s.s3-website.ap-south-1.amazonaws.com", bucket)
	return staticSiteURL, nil
}

func (s *deployService) uploadFile(svc *s3.Client, path, bucket string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	contentType := s.getContentType(path)
	key := strings.Join(strings.Split(path, string(os.PathSeparator))[3:], "/")

	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %v", path, err)
	}

	return nil
}

func (s *deployService) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
	}

	if contentType, ok := mimeTypes[ext]; ok {
		return contentType
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

// Function to add a DNS record using Cloudflare API
func (s *deployService) AddDNSRecord(url, projectName string) error {
	// Cloudflare API URL and token (replace with your actual API token and Zone ID)

	cfg := config.LoadConfig()
	apiToken := cfg.CloudflareApiToken
	zoneID := cfg.CloudflareZoneId
	fmt.Println(apiToken, zoneID, url, projectName, strings.Join(strings.Split(url, "/")[2:], "/"))
	// DNS Record data
	dnsRecord := map[string]interface{}{
		"type":    "CNAME",
		"name":    projectName + ".nymbus.xyz", // DNS name you want to add
		"content": strings.Join(strings.Split(url, "/")[2:], "/"),
		"ttl":     120,   // TTL in seconds
		"proxied": false, // Whether to enable Cloudflare proxying
	}

	// Serialize the DNS record to JSON
	recordData, err := json.Marshal(dnsRecord)
	if err != nil {
		return fmt.Errorf("error encoding DNS record: %w", err)
	}

	// Create a request to the Cloudflare API
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID), bytes.NewBuffer(recordData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set request headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add DNS record, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read and print the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	fmt.Println("DNS record added successfully:", string(bodyBytes))
	return nil
}

func (s *deployService) DeleteProject(projectID uint) error {
	// Get project details
	project, err := s.db.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project details: %w", err)
	}

	// Delete DNS record
	err = s.deleteDNSRecord(project.Name)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	// Delete S3 bucket contents and the bucket itself
	err = s.deleteS3Bucket(project.Name + ".nymbus.xyz")
	if err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}

	// Delete project from database
	err = s.db.DeleteProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project from database: %w", err)
	}

	return nil
}

func (s *deployService) deleteDNSRecord(projectName string) error {
	cfg := config.LoadConfig()
	apiToken := cfg.CloudflareApiToken
	zoneID := cfg.CloudflareZoneId

	// First, get the DNS record ID
	recordID, err := s.getDNSRecordID(apiToken, zoneID, projectName+".nymbus.xyz")
	if err != nil {
		return err
	}

	// Delete the DNS record
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete DNS record, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *deployService) getDNSRecordID(apiToken, zoneID, name string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Result) == 0 {
		return "", fmt.Errorf("no DNS record found for %s", name)
	}

	return result.Result[0].ID, nil
}

func (s *deployService) deleteS3Bucket(bucketName string) error {
	cfg, err := s.loadAWSConfig()
	if err != nil {
		return err
	}

	svc := s3.NewFromConfig(cfg)

	// Delete all objects in the bucket
	err = s.emptyS3Bucket(svc, bucketName)
	if err != nil {
		return err
	}

	// Delete the bucket
	_, err = svc.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete S3 bucket: %w", err)
	}

	return nil
}

func (s *deployService) emptyS3Bucket(svc *s3.Client, bucketName string) error {
	// Create a list objects paginator
	paginator := s3.NewListObjectsV2Paginator(svc, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	// Iterate through each page of objects
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		objectIds := make([]s3Types.ObjectIdentifier, len(page.Contents))
		for i, object := range page.Contents {
			objectIds[i] = s3Types.ObjectIdentifier{Key: object.Key}
		}

		// Delete the objects
		_, err = svc.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3Types.Delete{Objects: objectIds},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	return nil
}

func (s *deployService) CreateProject(project *domain.Project) (uint, error) {
	return s.db.CreateProject(project)
}

func (s *deployService) GetProject(id uint) (*domain.Project, error) {
	return s.db.GetProject(id)
}
