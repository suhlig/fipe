# `fly vipe`

Translates a Concourse pipeline URL like `https://concourse.example.com/teams/main/pipelines/ship-it?vars.product=%22foobar%22`

to

```command
fly -t $target get-pipeline --pipeline $pipeline/account:$account \
  | vipe \
  | fly -t $target set-pipeline --pipeline $pipeline --instance-var account=$account -c -
```
