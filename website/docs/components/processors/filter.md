---
title: filter
type: processor
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     lib/processor/filter.go
-->


```yaml
filter:
  text:
    arg: ""
    operator: equals_cs
    part: 0
  type: text
```

Tests each message batch against a condition, if the condition fails then the
batch is dropped. You can find a [full list of conditions here](/docs/components/conditions/about).

In order to filter individual messages of a batch use the
[`filter_parts`](/docs/components/processors/filter_parts) processor.


