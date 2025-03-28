# `fipe`

`fly + vipe = fipe`

Gets the configuration of a Concourse pipeline URL like `https://concourse.example.com/teams/main/pipelines/ship-it?vars.product=%22foobar%22&features=%22something%22`, pipes it into `vipe`, and sets the pipeline with the edited result.

# Requirements

It needs `fly` and `vipe`.

# FAQ

## Which editor is being used?

This program invokes `vipe`, which respects the `VISUAL` and `EDITOR` environment variables. So if you prefer `code` as an editor, run `fipe` as follows:

```command
$ VISUAL=`code -w` fipe "https://concourse.example.com/teams/main/pipelines/ship-it?vars.product=%22foobar%22&features=%22something%22"
```
