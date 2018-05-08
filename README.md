# AWS CloudWatch Logs exporter

Export AWS CloudWatch Logs to Prometheus (as labels and counts)

```bash
make
./bin/awslogs-exporter --aws.region="${AWS_REGION}"
```

## Notes:

* This exporter will listen by default on the port `9223`
* Requires AWS credentials or permission from an EC2 instance
* You can use the following IAM policy to grant required permissions:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "logs:DescribeLogGroups",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```


## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| awslogs_up | Was the last query of AWS Logs successful | region |
| TBD | TBD | TBD |

## Flags

* `aws.region`: The AWS region to get metrics from
* `aws.log-prefix`: Filter prefix log group names to search
* `aws.log-history`: Number of seconds to search for previous log events (default 1-hour)
  - AWS returns up to 1MB of oldest-to-newest logs, so high history values could drop most recent events
* `aws.log-json-format`: Converts line-by-line JSON log messages into pretty format
  - Use [`text/template`](https://golang.org/pkg/text/template/) formatting (i.e. `'{{.name}}: {{.message}}'`). Non-JSON lines will be printed normally.
* `debug`: Run exporter in debug mode
* `web.listen-address`: Address to listen on (default ":9223")
* `web.telemetry-path`: The path where metrics will be exposed (default "/metrics")

## Docker

You can deploy this exporter using the [houserater/awslogs-exporter](https://hub.docker.com/r/houserater/awslogs-exporter/) Docker image.

Note: Requires AWS credentials or permission from an EC2 instance, for example you can pass the env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options

For example:

```bash
docker pull houserater/awslogs-exporter
docker run -d -p 9223:9223 houserater/awslogs-exporter -aws.region="us-east-1"
```
