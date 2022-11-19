## svs: The StateVectorSync Protocol

<div align="center">

[**API Documentation**](https://pkg.go.dev/github.com/justincpresley/ndn-sync/pkg/svs) | | [**Examples**](/examples/svs/README.md)

</div>


### Helpful Links:
* [Technical Report](https://named-data.net/wp-content/uploads/2021/07/ndn-0073-r2-SVS.pdf)
* [Specification](https://named-data.github.io/StateVectorSync/)
* [Reference Implementation](https://github.com/named-data/ndn-svs)


### Sync Aspects:
```
Dataset Representation: VectorClock
Communication Model:    Push Notification
Dataset Range:          Full-data
Multicast Usage:        Yes
Data Naming:            Sequential
Packet Delivery:        Out-of-order
Strengths:              Resilient, Low Latency
Weaknesses:             Scalability, Naming
```


### Production Differences:
The Production branch **is** compatible with the Specification branch.

Differences:
* StateVectors are ordered by Latest Entries.
