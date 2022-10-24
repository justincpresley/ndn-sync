<div align="center">

![Visual](/docs/README_VISUAL.png)

![Tests](https://img.shields.io/github/workflow/status/justincpresley/ndn-sync/Tests?label=Tests)
![CodeFactor](https://img.shields.io/codefactor/grade/github/justincpresley/ndn-sync?label=CodeFactor)
![Language](https://img.shields.io/github/go-mod/go-version/justincpresley/ndn-sync)
![Version](https://img.shields.io/github/v/tag/justincpresley/ndn-sync?label=Latest%20version)
![License](https://img.shields.io/github/license/justincpresley/ndn-sync?label=License)

</div>

***ndn-sync*** is a [Go](https://go.dev/) library implementing [Named Data Networking](https://named-data.net/) (NDN) Distributed Dataset Synchronization '*Sync*' Protocols that can be used to write various real-time NDN Applications.

The goal of '*Sync*' is to inform others about updates in a dataset and/or to learn
about newly published data, effectively synchronizing data in a group.
***ndn-sync*** welcomes both newcomers and experts.

***ndn-sync*** is implemented using the NDN library [go-ndn](https://github.com/zjkmxy/go-ndn).


## Branches

***ndn-sync*** contains two main branches with their differences described below:

* [***production***](https://github.com/justincpresley/ndn-sync/tree/production): The master branch which holds Syncs along with any modifications to make them more stable/usable for applications. This branch is actively being served as a Go package.
* [***specification***](https://github.com/justincpresley/ndn-sync/tree/specification): The side branch which holds Syncs in their original form according to their technical specification.


## Syncs

There are many Syncs!

Ones that are being used in applications, others that are currently experiments,
and some that have yet to be discovered. ***ndn-sync*** gladly accepts any
kind of Sync protocol with a slight bias towards new and/or stable Syncs.

This [Sync Survey](https://named-data.net/wp-content/uploads/2021/05/ndn-0053-2-sync-survey.pdf)
describes many of the Syncs that are currently known and their unique differences. It is a recommended read.

***ndn-sync*** has the following Syncs implemented:

* **svs: StateVectorSync**

## Usage

In order to utilize ***ndn-sync***, two prerequisites are needed: [The NDN Forwarding Daemon (NFD)](https://named-data.net/doc/NFD/current/INSTALL.html) and [The Go Programming Language](https://go.dev/dl/).

It is highly recommended that you check out our examples. Sometimes, seeing the Syncs in action can give you ideas and help you in understanding what the Syncs provide.

***ndn-sync***'s API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/justincpresley/ndn-sync).
