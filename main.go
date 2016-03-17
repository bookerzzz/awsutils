package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
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
	ID                string `json:"Id,omitempty"`
	Status            string `json:"Status,omitempty"`
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

// DistSummary for summary output of a cloudfront distribution
type DistSummary struct {
	ID     string
	Domain string
	Alias  string
	Origin string
	Status string
}

// ByAlias implements sort.Interface for []DistSummary based on
// the Alias field.
type ByAlias []DistSummary

func (a ByAlias) Len() int           { return len(a) }
func (a ByAlias) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAlias) Less(i, j int) bool { return a[i].Alias < a[j].Alias }

// ByOrigin implements sort.Interface for []DistSummary based on
// the Alias field.
type ByOrigin []DistSummary

func (a ByOrigin) Len() int           { return len(a) }
func (a ByOrigin) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOrigin) Less(i, j int) bool { return a[i].Origin < a[j].Origin }

// ByStatus implements sort.Interface for []DistSummary based on
// the Alias field.
type ByStatus []DistSummary

func (a ByStatus) Len() int           { return len(a) }
func (a ByStatus) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStatus) Less(i, j int) bool { return a[i].Status < a[j].Status }

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
	var app *cli.App
	app = cli.NewApp()
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

							fn := fmt.Sprintf("%s.json", d.Origins.Items[0]["Id"])
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
				{
					Name:  "dists",
					Usage: "summarize the cloudfront distribution configurations",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "order-by",
							Value: "alias",
							Usage: "sort results on alias|origin|status",
						},
						cli.BoolFlag{
							Name:  "csv",
							Usage: "output as csv",
						},
					},
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

						var list []DistSummary
						for _, d := range r.Distributionlist.Items {
							id := d.ID
							distDomain := d.DomainName
							status := d.Status
							origin := ""
							alias := ""

							if len(d.Origins.Items) > 0 {
								for _, o := range d.Origins.Items {
									if op, ok := o["OriginPath"].(string); ok && op == "" {
										origin = o["DomainName"].(string)
									}
								}
							}
							if len(d.Aliases.Items) > 0 {
								alias = strings.Join(d.Aliases.Items, ",")
							}

							list = append(list, DistSummary{ID: id, Domain: distDomain, Alias: alias, Origin: origin, Status: status})
						}

						switch c.String("order-by") {
						case "origin":
							sort.Sort(ByOrigin(list))
						case "status":
							sort.Sort(ByStatus(list))
						case "alias":
							sort.Sort(ByAlias(list))
						default:
							fmt.Println("Unrecognised value for", c.String("order-by"), ". Sorting by alias instead.")
							sort.Sort(ByAlias(list))
						}

						if c.Bool("csv") {
							fmt.Println("ID;Domain;Alias;Origin;Status")
							for _, s := range list {
								fmt.Printf(`"%s";"%s";"%s";"%s";"%s"`, s.ID, s.Domain, s.Alias, s.Origin, s.Status)
								fmt.Println()
							}
						} else {
							w := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
							fmt.Fprint(w, "ID\tDomain\tAlias\tOrigin\tStatus\n")
							for _, s := range list {
								fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Domain, s.Alias, s.Origin, s.Status)
							}
							w.Flush()
						}
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
