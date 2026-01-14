# Introduction

Geoproxy is a TCP only proxy that uses the free ip-api.com json api to accept or block connections to specified backends based on IP Geolocation data.

Responses from ip-api.com are cached in RAM for 24 hours to try to cut down on requests as much as possible.  The cache is set to 10K entries by default and when it fills up, it will use an LRU algorithm to delete older entries.

I have recently added Proxy Protocol support so you can preserve the source IP to things that support proxy protocol.

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
  -lru int
    	size of the IP address LRU cache (default 10000)

```

# Sample Config

```
apiKey: xxxxx
servers:
  - listenIP: "0.0.0.0"
    listenPort: "22"
    backendIP: "192.168.6.1"
    backendPort: "22"
    daysOfWeek: ["Mon", "Tue", "Wed", "Thu", "Fri"]
    startDate: "2024-09-01"
    endDate: "2024-09-30"
    startTime: "08:00"
    endTime: "18:00"
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
    sendProxyProtocol: true
    recvProxyProtocol: true
```

Note: `daysOfWeek` cannot be combined with `startDate`/`endDate` in the same server block.

# Limitations

* As of 8/2024, ip-api throttles you to 45 free requests per minute.  I cache responses, but this is only going to get you so far.
* I don't pay for ip-api.  I added best effort support for it.  If it doesn't work, sorry in advance.
* I wouldn't call this well tested, but it works for my personal use case.
* I have tested proxy protocol with ipv4.
* I use Accept for TCP connections, so there are likely scaling limits.
* I don't fork anything and this doesn't run as a daemon.  That can be a problem if you don't like running as root for ports under 1024.
* I think IPv6 works ok, but I don't have IPv6 currently to test it out.
* I'm an ok Go dev, but if you know how to do something better, let me know.
* No udp support currently.  The only UDP things I care about either need to be locked to specific IPs, or need to run wide open.

# TODO

* Proxy protocol support.
