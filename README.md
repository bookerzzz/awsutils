# AWSUtils

Utililities to aid in the automation of Amazon Web Services configurations through their API.

## Assumptions

This utility makes use of the AWS command line interface and assumes it has been correctly configured. For
more assistance with this you can refer to http://docs.aws.amazon.com/cli/latest/userguide/installing.html#install-bundle-other-os
and http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html

## Available commands

* iam certs - List installed certificates and statuses
* cloudfront export-configs - Export cloudfront distribution configs (as json) to current working directory
