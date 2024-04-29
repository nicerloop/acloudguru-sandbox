# A Cloud Guru Sandbox

Command-line tool to manage A Cloud Guru sandboxes and configure credentials

## Usage

The A Cloud Guru credentials are retrieved from git credentials store using `git credential fill`.

The A Cloud Guru sandbox is started or stopped according to the requested target, and the credentials are configured for use by the local command line tools and SDKs.

```
acloudguru-sandbox <current|stop|aws|azure|gcloud> [-auth=<guru|google>][-rod=...]
```

Several authentications providers are supported with the optional parameter `[-auth=<guru|google>]`:

- guru: The default one, a cloud guru account
- google: Use authentification with a google account

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

## How to install

You can directly use the [released binaries](https://github.com/nicerloop/acloudguru-sandbox/releases) or use a package manager.

### Linux Requirements

[Rod](https://go-rod.github.io) can use an already installed browser or automatically download and use a statically versioned chromium browser. In case of Linux, the OS need to have [those dependency installed](https://github.com/go-rod/rod/blob/main/lib/docker/Dockerfile):
```
    apt-get update
    apt-get install --no-install-recommends -y \
    libnss3 \
    libxss1 \
    libasound2 \
    libxtst6 \
    libgtk-3-0 \
    libgbm1 \
    ca-certificates \
    fonts-liberation fonts-noto-color-emoji fonts-noto-cjk \
    tzdata \
    dumb-init \
    xvfb
```

### Windows with Scoop

Install [Scoop](https://scoop.sh/):

```
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
```

Add [nicerloop/nicerloop bucket](https://scoop.sh/#/apps?q=%22https%3A%2F%2Fgithub.com%2Fnicerloop%2Fscoop-nicerloop%22&o=false):

```
scoop bucket add nicerloop https://github.com/nicerloop/scoop-nicerloop
```

Install acloudguru-sandbox:

```
scoop install nicerloop/acloudguru-sandbox
```

## Similar works and inspiration

https://github.com/josephedward/gosandbox
