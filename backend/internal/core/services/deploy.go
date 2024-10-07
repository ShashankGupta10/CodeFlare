package services

import (
	"bytes"
	"codeflare/internal/adapters/repository"
	"codeflare/internal/core/ports"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-git/go-git/v5"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type deployService struct {
	db *repository.PGStore
}

func NewDeployService(db *repository.PGStore) port.DeployService {
	return &deployService{db: db}
}

func GetFilePaths(repoPath string) ([]string, error) {

	// DFS in development??? 😱
	var filePaths []string
	q := []string{repoPath + "/dist"}

	for len(q) > 0 {
		curr := q[0]
		q = q[1:]

		data, err := os.ReadDir(curr)
		if err != nil {
			return nil, err
		}

		for _, item := range data {
			fullpath := filepath.Join(curr, item.Name())
			if item.IsDir() {
				q = append(q, fullpath+"/")
			} else {
				filePaths = append(filePaths, fullpath)
			}
		}
	}
	return filePaths, nil
}

func (s *deployService) AlreadyDeployed(url string) (bool, error) {
	val, err := s.db.FindRepo(url)
	if err != nil {
		return false, err
	}
	return val, nil
}

func (s *deployService) ValidateURL(url string) error {
	parts := strings.Split(url, "/")
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	resp, err := http.Get("https://api.github.com/repos/" + owner + "/" + repo)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("repository not found")
	}

	return nil
}

func (s *deployService) CloneRepo(url string) (string, error) {
	repoName := strings.Split(url, "/")[4]
	destination := "./projects/" + repoName

	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL:      url,
		Progress: nil,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %v", err)
	}
	return destination, nil
}

func (s *deployService) BuildRepo(dir string) (string, error) {
	// First, install the project dependencies
	installCmd := exec.Command("npm", "install")
	installCmd.Dir = dir

	// Run npm install
	output, err := installCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to install dependencies: %s, error: %v", string(output), err)
	}
	fmt.Println("Project dependencies installed successfully")
	// Next, build the React project
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = dir

	// Run npm run build
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build project: %s, error: %v", string(buildOutput), err)
	}
	fmt.Println("Project built successfully")
	// Return success message or the build output
	fmt.Println(string(buildOutput))
	return string(buildOutput), nil
}

func (s *deployService) UploadToS3(dir string) (string, error) {
	// Load AWS configuration
	fmt.Println("In upload func", dir)
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("ap-south-1"))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	fmt.Println("config loaded")

	// Create S3 service client
	svc := s3.NewFromConfig(cfg)

	bucket := "swipe-assignment" // Replace with your desired bucket name
	fmt.Println("BUCKET NAME GIVEN")
	// Check if the bucket exists, if not, create it
	_, err = svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		fmt.Println("HER", err)
		// Bucket doesn't exist, create it
		_, createBucketErr := svc.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
				LocationConstraint: "ap-south-1",
			},
		})
		if createBucketErr != nil {
			fmt.Println("CREATE BUCK ERROR", createBucketErr)
			return "", fmt.Errorf("failed to create bucket: %v", createBucketErr)
		}

		// Disable "Block all public access" settings
		_, accessErr := svc.PutPublicAccessBlock(context.TODO(), &s3.PutPublicAccessBlockInput{
			Bucket: aws.String(bucket),
			PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
				BlockPublicAcls:       aws.Bool(false),
				IgnorePublicAcls:      aws.Bool(false),
				BlockPublicPolicy:     aws.Bool(false),
				RestrictPublicBuckets: aws.Bool(false),
			},
		})
		if accessErr != nil {
			return "", fmt.Errorf("failed to disable public access block: %v", err)
		}

		// Enable static site hosting
		_, putBucketErr := svc.PutBucketWebsite(context.TODO(), &s3.PutBucketWebsiteInput{
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
		if putBucketErr != nil {
			fmt.Println("PUT BUCK ERROR", putBucketErr)
			return "", fmt.Errorf("failed to enable static site hosting: %v", err)
		}

		// Set public access policy
		publicPolicy := `{
					"Version": "2012-10-17",
					"Statement": [
						{
							"Sid": "Stmt1405592139000",
							"Effect": "Allow",
							"Principal": "*",
							"Action": "s3:GetObject",
							"Resource": [
								"arn:aws:s3:::` + bucket + `/*",
								"arn:aws:s3:::` + bucket + `"
							]
						}
					]
				}`

		_, putBucketPolicyErr := svc.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
			Bucket: aws.String(bucket),
			Policy: aws.String(publicPolicy),
		})
		if putBucketPolicyErr != nil {
			fmt.Println("HELLO BRO", putBucketPolicyErr)
			return "", fmt.Errorf("failed to set public bucket policy: %v", err)
		}
	}

	fmt.Println("Bucket ready")

	// Iterate through all files in the build directory and upload them
	fmt.Println(dir)
	files, filesPathErr := GetFilePaths(dir)
	if filesPathErr != nil {
		fmt.Println(filesPathErr)
		return "", nil
	}

	fmt.Println(files)
	for _, path := range files {
		file, OpenFileErr := os.Open(path)
		if OpenFileErr != nil {
			return "", OpenFileErr
		} else {
			defer file.Close()
			ext := strings.ToLower(filepath.Ext(path))
			mimeTypes := map[string]string{
				".html": "text/html",
				".css":  "text/css",
				".js":   "application/javascript",
			}

			// Find the correct ContentType from the map, fallback to default
			contentType := mimeTypes[ext]
			if contentType == "" {
				// Use the "mime" package as a fallback to guess the ContentType
				contentType = mime.TypeByExtension(ext)
				if contentType == "" {
					contentType = "application/octet-stream" // Default to binary data
				}
			}
			_, PubObjErr := svc.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket:      aws.String(bucket),
				Key:         aws.String(strings.Join(strings.Split(path, "\\")[3:], "/")),
				Body:        file,
				ContentType: &contentType,
			})

			if PubObjErr != nil {
				return "", PubObjErr
			}
		}

	}

	fmt.Println("Files uploaded")

	// Return the bucket URL and the static site URL
	bucketURL := fmt.Sprintf("https://%s.s3.amazonaws.com/", bucket)
	staticSiteURL := fmt.Sprintf("http://%s.s3-website.ap-south-1.amazonaws.com", bucket)

	fmt.Print(staticSiteURL, bucketURL)

	return staticSiteURL, nil
}

// Function to add a DNS record using Cloudflare API
func (s *deployService) AddDNSRecord(url string) error {
	// Cloudflare API URL and token (replace with your actual API token and Zone ID)
	apiToken := "your-cloudflare-api-token"
	zoneID := "your-cloudflare-zone-id"

	// DNS Record data
	dnsRecord := map[string]interface{}{
		"type":    "A",                        // Adjust based on record type (e.g., A, CNAME, MX)
		"name":    "subdomain.yourdomain.com", // DNS name you want to add
		"content": "192.0.2.1",                // The IP or content of the DNS record
		"ttl":     120,                        // TTL in seconds
		"proxied": false,                      // Whether to enable Cloudflare proxying
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

	// Read the response and check for success
	body := new(strings.Builder)
	// _, _ = body.(resp.Body)

	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("failed to add DNS record: %s", body.String())
	// }

	fmt.Print(body)

	fmt.Println("DNS record added successfully:", body.String())
	return nil
}
