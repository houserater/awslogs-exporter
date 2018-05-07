# AWS Logs exporter [![Build Status](https://travis-ci.org/houserater/awslogs-exporter.svg?branch=master)](https://travis-ci.org/houserater/awslogs-exporter)

Export AWS CloudWatch Logs to Prometheus

```bash
make
./bin/awslogs-exporter --aws.region="${AWS_REGION}"
```

## Notes:

* This exporter will listen by default on the port `9222`
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
                "logs:DescribeLogStreams",
                "logs:GetLogEvents"
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
* `aws.group-filter`: Regex used to filter the log group names, if doesn't match the group is ignored (default ".*")
* `debug`: Run exporter in debug mode
* `web.listen-address`: Address to listen on (default ":9222")
* `web.telemetry-path`: The path where metrics will be exposed (default "/metrics")

## Docker

You can deploy this exporter using the [houserater/awslogs-exporter](https://hub.docker.com/r/houserater/awslogs-exporter/) Docker image.

Note: Requires AWS credentials or permission from an EC2 instance, for example you can pass the env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options

For example:

```bash
docker pull houserater/awslogs-exporter
docker run -d -p 9222:9222 houserater/awslogs-exporter -aws.region="us-east-1"
```
