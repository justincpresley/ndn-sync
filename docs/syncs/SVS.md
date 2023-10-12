# svs: The StateVectorSync Protocol

<div align="center">

[**API Documentation**](https://pkg.go.dev/github.com/justincpresley/ndn-sync/pkg/svs) | | [**Examples**](/examples/svs/README.md)

</div>

> **Warning**
> This Sync is currently vulnerable to many attacks due to security not being 'filled-in' and should not be used in a production environment until this notice is removed.


### Helpful Links:
* [Technical Report](https://named-data.net/wp-content/uploads/2021/07/ndn-0073-r2-SVS.pdf)
* [Specification](https://named-data.github.io/StateVectorSync/)
* [Reference Implementation](https://github.com/named-data/ndn-svs)
* [Scalability Paper](https://dl.acm.org/doi/pdf/10.1145/3517212.3559485)


### Natural Aspects:
```
Dataset Representation: VectorClock
Communication Model:    Push Notification
Dataset Range:          Full-data
Dataset Roles:          No Separation or Definition
Multicast Usage:        Yes
Long-lived Interests:   No
Data Naming:            Sequential
Packet Delivery:        Out-of-order
Strengths:              Resilient, Low Latency
Weaknesses:             Scalability, Naming
Additional Notes:       Key Establishment for Group
```


### Production Differences:
The [Production branch](https://github.com/justincpresley/ndn-sync/tree/production) **is** compatible with the [Specification branch](https://github.com/justincpresley/ndn-sync/tree/specification) **if** using FormalEncoding.

Differences:
* New Sync types: HealthSync, SharedSync.
* StateVectors are ordered via Latest Entries First.
* Optimized Informal StateVector Encoding
