#!/bin/sh

/opt/bitnami/kafka/bin/kafka-topics.sh --create --topic xdd --bootstrap-server localhost:9092
/opt/bitnami/kafka/bin/kafka-topics.sh --describe --topic xdd --bootstrap-server localhost:9092
