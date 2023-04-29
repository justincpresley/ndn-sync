<div align="center">

![Visual](/docs/README_VISUAL.png)

![Test](https://img.shields.io/github/actions/workflow/status/justincpresley/ndn-sync/test.yaml?branch=production&label=Test)
![CodeQL](https://img.shields.io/github/actions/workflow/status/justincpresley/ndn-sync/codeql.yml?branch=production&label=CodeQL)
![CodeFactor](https://img.shields.io/codefactor/grade/github/justincpresley/ndn-sync/production?label=CodeFactor)
![Language](https://img.shields.io/github/go-mod/go-version/justincpresley/ndn-sync/production?label=Go)
![Version](https://img.shields.io/github/v/tag/justincpresley/ndn-sync?label=Latest%20version)
![Commits](https://img.shields.io/github/commits-since/justincpresley/ndn-sync/latest/production?label=Unreleased%20commits)
![License](https://img.shields.io/github/license/justincpresley/ndn-sync?label=License)

</div>

***ndn-sync*** is a [Go](https://go.dev/) library implementing [Named Data Networking](https://named-data.net/) (NDN) Distributed Dataset Synchronization '*Sync*' Protocols that can be used to write various real-time NDN Applications.

The goal of '*Sync*' is to inform others about updates in a dataset and/or to learn
about newly published data, effectively synchronizing data in a group.
***ndn-sync*** welcomes both newcomers and experts of NDN.

***ndn-sync*** is implemented using the NDN library [go-ndn](https://github.com/zjkmxy/go-ndn).


## Branches

***ndn-sync*** contains two main branches with their differences described below:

* [**production**](https://github.com/justincpresley/ndn-sync/tree/production): The master branch which holds Syncs along with any modifications to make them more stable/usable for applications. This branch is actively being served as a Go package.
* [**specification**](https://github.com/justincpresley/ndn-sync/tree/specification): The side branch which holds Syncs in their original form according to their technical specification.


## Usage

Before you utilize ***ndn-sync*** or try any of its examples, please ensure that you have the necessary [prerequisites](/docs/INSTALL.md). It will take but a few minutes!

***ndn-sync*** is a library containing multiple modules (different Syncs), each with individual functionality and use.

It is highly recommended that you check out the examples. Sometimes, seeing the Syncs in action can give you ideas and help you in understanding what the Syncs provide.


## Syncs

There are many Syncs!

Ones that are being used in applications, others that are currently experiments,
and some that have yet to be discovered. ***ndn-sync*** gladly accepts any
kind of Sync protocol with a slight bias towards new and/or stable Syncs.

This [Sync Survey](https://named-data.net/wp-content/uploads/2021/05/ndn-0053-2-sync-survey.pdf)
describes many of the Syncs that are currently known and their unique differences. It is a recommended read.

***ndn-sync*** has the following Syncs implemented:

* `svs` - **StateVectorSync**: [Details](/docs/syncs/SVS.md) | [API Documentation](https://pkg.go.dev/github.com/justincpresley/ndn-sync/pkg/svs) | [Examples](/examples/svs/README.md)


## Getting Involved

The most effortless way you can contribute to ***ndn-sync*** is to simply have discussions surrounding ***ndn-sync***. [You can do so here!](https://github.com/justincpresley/ndn-sync/discussions)

In addition, ***ndn-sync*** has more practical ways to get involved: [Issues](https://github.com/justincpresley/ndn-sync/issues) and [Pull Requests](https://github.com/justincpresley/ndn-sync/pulls).

## License

***ndn-sync*** is an open source project licensed under ISC. See LICENSE.md for more information.
