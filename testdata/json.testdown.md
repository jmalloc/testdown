# JSON

This document describes a test of a JSON pretty-printer.

The input is shown below:

```json
{ "one": 1, "two": 2 }
```

If the pretty-printer is functioning correctly, it should produce
JSON that is indented and and separated onto new lines, as follows:

```json testdown
{
  "one": 1,
  "two": 2
}
```
