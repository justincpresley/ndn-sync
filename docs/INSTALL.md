## Installation

In order to utilize ***ndn-sync***, two prerequisites are needed. The first of which should be most apparent, you need Go! In addition, you need a compatible NDN Forwarder (think of it like a router).


### Getting Go

There are tons of good tutorials to download / install [The Go Programming Language](https://go.dev/). I find [this one](https://www.digitalocean.com/community/tutorials/how-to-install-go-on-ubuntu-20-04) for Linux to be quick and simple. Nevertheless, [this is the main Go download page](https://go.dev/dl/).

It is suggested that you use the latest version of Go to ensure compatibility with ***ndn-sync***.


### Choosing your Forwarder

There are many forwarders available:
* **NFD**: [The NDN Forwarding Daemon](https://github.com/named-data/NFD)
* **YaNFD**: [Yet-another NDN Forwarding Daemon](https://github.com/named-data/YaNFD)
* **NDN-DPDK**: [NDN Data Plane Development Kit](https://github.com/usnistgov/ndn-dpdk)

However, ***ndn-sync*** currently only supports **NFD** until more support is added to [go-ndn](https://github.com/zjkmxy/go-ndn). While **NFD** does have [nice documentation](https://named-data.net/doc/NFD/current/INSTALL.html) to install, I prefer to use [yoursunny's nightly APT repository](https://yoursunny.com/t/2021/NFD-nightly-usage/).

You can test that **NFD** is properly installed by running `nfd-start` and than `nfd-status`.


### Using this Library

After both prerequisites are installed, you can import Syncs and use them in your applications! Of course, Go will complain if you do not fetch the library via `go get`.

If you `git clone` ***ndn-sync***, you can run the following instead and run the examples with the source that you have.
```
go get -u -v -f all
```
