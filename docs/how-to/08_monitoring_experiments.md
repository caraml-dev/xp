# Monitoring Experiments

The Experiments may be monitored on Prometheus. This includes the following metrics.

- Performance of the app (latency, throughput, error rate, etc.) and various components (such as calls to DB)
- Resource utilization
- Experiment Lookup
- Treatment assignment (Matching, non-matching experiments)

## Treatment logs

Treatment request and response log are available to be written to BigQuery or Kafka.
