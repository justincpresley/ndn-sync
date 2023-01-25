# svs: The StateVectorSync Protocol

<div align="center">

[**API Documentation**](https://pkg.go.dev/github.com/justincpresley/ndn-sync/pkg/svs) | | [**Examples**](/examples/svs/README.md)

</div>


### Helpful Links:
* [Technical Report](https://named-data.net/wp-content/uploads/2021/07/ndn-0073-r2-SVS.pdf)
* [Specification](https://named-data.github.io/StateVectorSync/)
* [Reference Implementation](https://github.com/named-data/ndn-svs)


### Natural Sync Aspects:
```
Dataset Representation: VectorClock
Communication Model:    Push Notification
Dataset Range:          Full-data
Multicast Usage:        Yes
Long-lived Interests:   No
Data Naming:            Sequential
Packet Delivery:        Out-of-order
Strengths:              Resilient, Low Latency
Weaknesses:             Scalability, Naming
```


### Production Differences:
The Production branch **is** compatible with the Specification branch **if** using FormalEncoding.

Differences:
* New Sync types: HealthSync, SharedSync.
* StateVectors are ordered via Latest Entries (descending in freshness).
* Optimized Informal StateVector Encoding
