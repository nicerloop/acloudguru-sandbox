# A Cloud Guru Sandbox

Command-line tool to manage A Cloud Guru sandboxes and configure credentials

## Usage

The A Cloud Guru credentials are retrieved from git credentials store using `git dredentials fill`.

The A Cloud Guru sandbox is started or stopped according to the requested target, and the credentials are configured for use by the local command line tools and SDKs.

```
acloudguru-sandbox <current|stop|aws|azure|gcloud> [-rod=...]
```

## How does it work

The code to get credentials from git originates from [gitauth](https://pkg.go.dev/golang.org/x/tools/cmd/auth/gitauth).

[Rod](https://go-rod.github.io) is used to automate the required actions on [A Cloud Guru website](https://learn.acloud.guru/cloud-playground/cloud-sandboxes).

You can see the actual interaction through the browser using [Rod confiduration parameters](https://go-rod.github.io/#/get-started/README?id=slow-motion-and-visual-trace):

```
-rod=show,slow=1s,trace
```

## Using with a proxy server

Add go-rod proxy parameters:
```
-rod=proxy=http://<host>:<port>
```

## Similar works and inspiration

https://github.com/josephedward/gosandbox
