# Introduction to XP

XP support designing and managing experiment configurations in a safe and holistic manner. At run time, these configurations can be used (within the Turing router, or externally) to run the experiments and generate treatments. The experiments can be run either deterministically (A/B Experiments) or as a function of time (Switchback Experiments), or a combination of both (Randomized Switchbacks).

## Features

- **Reliable** - Inherent fault-detection rules help create experiments without conflicts
- **Customizable** - Every service has unique requirements. XP allows for defining flexible segmentation and experiment validation rules.
- **Fast** - 99p server-side latency (excluding the network latency between the calling service and XP) averages around 1 ms
- **Observable** - Resource utilization, treatment assignment and performance metrics are available on Prometheus
