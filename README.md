# Introduction

Geoproxy is a TCP only proxy that uses the free ip-api.com json api to accept or block connections to specified backends based on IP Geolocation data.

Responses from ip-api.com are cached in RAM for 24 hours to try to cut down on requests as much as possible.

# Installation

git clone this repository and run go build.  You'll want to set up a systemd unit file to start/stop/restart on errors.

# Useage

```
Usage of ./geoproxy:
  -action string
        iptables action to take on blocked IPs. Default is DROP. (default "DROP")
  -config string
        Path to the configuration file (default "geoproxy.yaml")
  -continue
        allow connections through on ipapi errors
  -ipapi string
        ipapi endpoint (default "http://ip-api.com/json/")
  -iptables string
        add rejected IPs to the specified iptables chain
```

# Sample Config

```
servers:
  - listenIP: "0.0.0.0"
    listenPort: "22"
    backendIP: "192.168.6.1"
    backendPort: "22"
    allowedCountries:
      - "US"
    allowedRegions:
      - "CA"
  - listenIp: "0.0.0.0"
    listenPOrt: "443"
    backendIP: "192.168.5.2"
    backendPort: "443"
    deniedCountries:
       - "RU"
```

# Limitations

* ip-api.com heavily throttles free requests.  
* I dont't want to pay for ip-api.com, so I don't have support for their auth tokens.
* I wouldn't call this well tested, but it works for my personal use case.
* I use Accept for TCP connections, so there are likely scaling limits.
* I don't fork anything and this doesn't run as a daemon.  That can be a problem if you don't like running as root for ports under 1024.

# TODO

* Proxy protocol support.
