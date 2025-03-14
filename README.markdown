# `fly vipe`

Translates a Concourse pipeline URL like `https://concourse.example.com/teams/main/pipelines/ship-it?vars.product=%22foobar%22&features=%22something%22`

to

```command
fly -t $target get-pipeline --pipeline $pipeline/product:foobar \
  | vipe \
  | fly -t $target set-pipeline --pipeline $pipeline --instance-var product=foobar --instance-var features=something -c -
```

I alias it as `fipe`.
