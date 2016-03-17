# AWSUtils

Utililities to aid in the automation of Amazon Web Services configurations through their API.

## Install
```bash
go install github.com/bookerzzz/awsutils
```

## Assumptions

This utility makes use of the AWS command line interface and assumes it has been correctly configured. For
more assistance with this you can refer to http://docs.aws.amazon.com/cli/latest/userguide/installing.html#install-bundle-other-os
and http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

## Available commands (so far)

* **iam certs** - List installed certificates and statuses
* **cloudfront export-configs** - Export cloudfront distribution configs (as json) to current working directory
* **awsutils cloudfront dists** - List configured cloudfront distributions and statuses

## Usage
```bash
awsutils iam certs
# prints list of certification available on AWS

awsutils cloudfront export-configs
# exports distribution configs from AWS CloudFront service into json
# files (named {dist-origin-id}.json) into current working directory.

awsutils cloudfront dists [--order-by=alias|origin|status] [--csv]
# prints list of distributions configured on AWS sorted on alias
# or optionally sorted on another supported value. Set output mode
# --csv if needed
```
