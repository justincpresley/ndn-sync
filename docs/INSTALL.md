# Installation

In order to utilize ***ndn-sync***, two prerequisites are needed. The first of which should be most apparent, you need Go! In addition, you need a compatible NDN Forwarder (think of it like a router).


## Getting Go

There are tons of good tutorials to download / install [The Go Programming Language](https://go.dev/). I find [this one](https://www.digitalocean.com/community/tutorials/how-to-install-go-on-ubuntu-20-04) for Linux to be quick and simple. Nevertheless, [this is the main Go download page](https://go.dev/dl/).

It is suggested that you use the latest version of Go to ensure compatibility with ***ndn-sync***.


## Choosing your Forwarder

There are many forwarders available:
* **NFD**: [The NDN Forwarding Daemon](https://github.com/named-data/NFD)
* **YaNFD**: [Yet-another NDN Forwarding Daemon](https://github.com/named-data/YaNFD)
* **NDN-DPDK**: [NDN Data Plane Development Kit](https://github.com/usnistgov/ndn-dpdk)

However, ***ndn-sync*** relies on [go-ndn](https://github.com/zjkmxy/go-ndn) for its forwarder support. As such, ***ndn-sync*** supports **NFD** and **YaNFD** only.

While **NFD** does have [nice documentation](https://named-data.net/doc/NFD/current/INSTALL.html) for installation, I prefer to use [yoursunny's nightly APT repository](https://yoursunny.com/t/2021/NFD-nightly-usage/).

You can test that **NFD** is properly installed by running `nfd-start` and then `nfd-status`. **YaNFD** does not directly install into *systemd* as a service. Instead, it must be self-started via `yanfd`.

## Using this Library

After both prerequisites are installed, you can import Syncs and use them in your applications! Of course, Go might complain if you do not retrieve the library via `go get`.

If you `git clone` ***ndn-sync***, you can simply run the examples out of the box and Go will automatically retrieve the needed dependencies.
