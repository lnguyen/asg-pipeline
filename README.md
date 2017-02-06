# ASG Pipeline

Creates Application Security Groups on Cloud Foundry based on folder structure


## Rules

By default any rules that are labeled with with one word or without `:` is considered global ASG and `org:space` is consider ASG that will be bound to that org and space.
