# headscale2hosts

This program connects to a [headscale](https://github.com/juanfont/headscale) instance and creates a [hosts](https://en.wikipedia.org/wiki/Host_file) file for each device in the network. This hosts file can be consumed by CoreDNS or put to `/etc/hosts`.


## Why?

I really love tailscale, but super-poopy VPN at work is very adamant about crapping all over my `/etc/resolv.conf` file, so I cannot use use MagicDNS from tailscale.

To solve this I use this program to create a hosts file and expose it to CoreDNS an dhave the dns records be public. This way I can use tailscale without any additional configuration. The downside is that the names of the devices are basically exposed to the internet.

## Usage

You can either run the program directly or use the supplied docker image. The program is configured  using the following environment variables:

| Name | Default | Description |
| ---- | ------- | ----------- |
| `HEADSCALE2HOSTS_SERVER_URL` | **REQUIRED** | The URL of the headscale server (same as `--auth-server` parameter you use to connect to headscale)|
| `HEADSCALE2HOSTS_API_KEY` | **REQUIRED** | The API key you use to connect to headscale |
| `HEADSCALE2HOSTS_NAMESPACE` | **REQUIRED** | The namespace to query from |
| `HEADSCALE2HOSTS_DOMAIN_SUFFIX` | `` | The domain suffix to add to each of the devices |
| `HEADSCALE2HOSTS_HOSTS_FILE` | `hosts` | The path to the hosts file to write |
| `HEADSCALE2HOSTS_CHECK_INTERVAL` | `1m` | How often to query headscale and regenerate the hosts file |

You can obtain the API key by running the following command on the coordinator:

```sh
$ headscale apikeys create -e 999999h
```

## Example `docker-compose.yml`

```yaml
version: "3"
services:
  headscale2hosts:
    image: ghcr.io/alufers/headscale2hosts:latest
    restart: unless-stopped
    volumes:
      - ./dnsconfig:/dnsconfig
    environment:
      HEADSCALE2HOSTS_SERVER_URL: "<headscale server url>"
      HEADSCALE2HOSTS_API_KEY: "<headscale api key>"
      HEADSCALE2HOSTS_NAMESPACE: "<headscale namespace>"
      HEADSCALE2HOSTS_DOMAIN_SUFFIX: "<domain suffix>"
      HEADSCALE2HOSTS_HOSTS_FILE: /dnsconfig/headscale.hosts
```
