<div align="center">

![Visual](/docs/README_VISUAL.png)

![Tests](https://img.shields.io/github/workflow/status/justincpresley/ndn-sync/Tests?label=Tests)
![Language](https://img.shields.io/github/go-mod/go-version/justincpresley/ndn-sync)
![Version](https://img.shields.io/github/v/tag/justincpresley/ndn-sync?label=Latest%20version)
![License](https://img.shields.io/github/license/justincpresley/ndn-sync?label=License)

</div>

**ndn-sync** is a Go library implementing Named Data Networking (NDN) Distributed
Dataset Synchronization ('*Sync*' for short) Protocols that can be used to write
various real-time NDN Applications. The goal of '*Sync*' is to inform others about
updates in a dataset and/or to learn about newly published data. **ndn-sync** welcomes
both newcomers and experts.


##Table of contents
<!--ts-->
   * [Branches](#branches)
   * [Syncs](#syncs)
<!--te-->


#Branches
**ndn-sync** is broken into two main branches with their differences described below:

* **production**: The master branch which holds Syncs along with modifications to make them more stable/usable for applications. This branch is actively being served as a Go package.
* **specification**: The side branch which holds Syncs in their original form according to their technical specification.


#Syncs
There are many Sync protocols! Ones that are being used in applications, others
that are currently experiments, and even some that have yet to be discovered.
**ndn-sync** gladly accepts any kind of Sync protocol with a slight bias towards
new and/or stable Syncs.

This [Sync Survey](https://named-data.net/wp-content/uploads/2021/05/ndn-0053-2-sync-survey.pdf)
describes many of the Syncs that are currently known and their unique differences. It is a highly
recommended read.

**ndn-sync** has the following Sync protocols implemented:

* **svs: StateVectorSync**
