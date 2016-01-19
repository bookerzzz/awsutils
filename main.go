package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
)

// ListServerCertificatesResponse model
type ListServerCertificatesResponse struct {
	ServerCertificateMetadataList []ServerCertificateMetadata
}

// ServerCertificateMetadata model
type ServerCertificateMetadata struct {
	ServerCertificateID   string `json:"ServerCertificateId"`
	ServerCertificateName string
	Expiration            time.Time
	Path                  string
	Arn                   string
	UploadDate            time.Time
}

// CloudFrontListDistributionsResponse model
type CloudFrontListDistributionsResponse struct {
	Distributionlist struct {
		Items []CloudFrontDistributionConfig
	}
}

// CloudFrontDistributionConfig model
type CloudFrontDistributionConfig struct {
	WebACLID          string `json:"WebACLId"`
	DefaultRootObject string
	PriceClass        string
	Enabled           bool
	CallerReference   string
	DomainName        string
	Comment           string
	CacheBehaviors    struct {
		Items    []map[string]interface{}
		Quantity int
	}
	Logging struct {
		Bucket         string
		Prefix         string
		Enabled        bool
		IncludeCookies bool
	}
	Origins struct {
		Quantity int
		Items    []map[string]interface{}
	}
	DefaultCacheBehavior struct {
		TargetOriginID       string `json:"TargetOriginId"`
		ViewerProtocolPolicy string
		MinTTL               int
		MaxTTL               int
		DefaultTTL           int
		SmoothStreaming      bool
		TrustedSigners       struct {
			Enabled  bool
			Quantity int
			Items    []string
		}
		ForwardedValues struct {
			QueryString bool
			Headers     struct {
				Items    []string
				Quantity int
			}
			Cookies struct {
				Forward string
			}
		}
		AllowedMethods struct {
			Items         []string
			Quantity      int
			CachedMethods struct {
				Items    []string
				Quantity int
			}
		}
	}
	ViewerCertificate struct {
		CloudFrontDefaultCertificate bool
		SSLSupportMethod             string
		MinimumProtocolVersion       string
		IAMCertificateID             string `json:"IAMCertificateId"`
	}
	CustomErrorResponses struct {
		Quantity int
		Items    []map[string]interface{}
	}
	Restrictions struct {
		GeoRestriction struct {
			RestrictionType string
			Quantity        int
		}
	}
	Aliases struct {
		Items    []string
		Quantity int
	}
}

// ResolveStatus resolve a certificates expiry status
func ResolveStatus(t time.Time) string {
	now := time.Now()
	if t.Unix() > now.AddDate(0, 1, 0).Unix() {
		return "OK"
	}
	if t.Unix() > now.Unix() {
		return "Expiring soon"
	}
	return "Expired"
}

func main() {
	app := cli.NewApp()
	app.Name = "awsutils"
	app.Usage = "automation of Amazon Web Services configurations through their API"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:  "iam",
			Usage: "use the AWS iam API",
			Subcommands: []cli.Command{
				{
					Name:  "certs",
					Usage: "summarise certificate configurations available in your AWS account",
					Action: func(c *cli.Context) {
						out, err := exec.Command("aws", "iam", "list-server-certificates").Output()
						if err != nil {
							fmt.Println(err.Error())
							return
						}

						var r ListServerCertificatesResponse
						err = json.Unmarshal(out, &r)
						if err != nil {
							fmt.Println(err.Error())
							return
						}

						w := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
						fmt.Fprint(w, "ID\tName\tStatus\n")
						for _, cm := range r.ServerCertificateMetadataList {
							fmt.Fprintf(w, "%s\t%s\t%s\n", cm.ServerCertificateID, cm.ServerCertificateName, ResolveStatus(cm.Expiration))
						}
						w.Flush()
						// fmt.Printf("%+v\n", r)
					},
				},
			},
		},
		{
			Name:    "cloudfront",
			Aliases: []string{"cf"},
			Usage:   "use the AWS cloudfront API",
			Subcommands: []cli.Command{
				{
					Name:  "export-configs",
					Usage: "export the cloudfront distribution configurations",
					Action: func(c *cli.Context) {
						out, err := exec.Command("aws", "cloudfront", "list-distributions").Output()
						if err != nil {
							fmt.Println(err.Error())
							return
						}

						var r CloudFrontListDistributionsResponse
						err = json.Unmarshal(out, &r)
						if err != nil {
							fmt.Println(err.Error())
							return
						}

						w := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
						fmt.Fprint(w, "DomainName\n")
						for _, d := range r.Distributionlist.Items {
							fmt.Fprintf(w, "%+v\n", d.Origins.Items[0]["DomainName"])
							json, _ := json.MarshalIndent(d, "", "  ")

							fn := fmt.Sprintf("%s.json", d.Origins.Items[0]["DomainName"])
							f, err := os.Create(fn)
							if err != nil {
								cwd, _ := os.Getwd()
								fmt.Printf("Unable to create file '%s' in '%s'", fn, cwd)
								continue
							}
							f.Write(json)
							f.Close()
						}
						w.Flush()
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
