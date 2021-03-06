---
title: try
type: processor
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     lib/processor/try.go
-->


```yaml
try: []
```

Behaves similarly to the [`for_each`](/docs/components/processors/for_each) processor, where a
list of child processors are applied to individual messages of a batch. However,
if a processor fails for a message then that message will skip all following
processors.

For example, with the following config:

``` yaml
- try:
  - type: foo
  - type: bar
  - type: baz
```

If the processor `foo` fails for a particular message, that message
will skip the processors `bar` and `baz`.

This processor is useful for when child processors depend on the successful
output of previous processors. This processor can be followed with a
[catch](/docs/components/processors/catch) processor for defining child processors to be applied
only to failed messages.

More information about error handing can be found [here](/docs/configuration/error_handling).


