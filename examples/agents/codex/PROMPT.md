# Sample Codex prompts

## Safe upgrade

```
Upgrade the web service on the prod host using shum. Dry-run first, summarize
the plan to me, then ask before applying. If shum returns any error code I
should know about (migration_warning, host_unreachable, rollback_failed),
stop and tell me.
```

## Inspection only

```
Show me what's currently deployed on the prod host. List registered hosts,
discover compose projects on prod, and show me the last 5 upgrade runs.
Use shum --json throughout.
```

## Roll back the latest run

```
Find the most recent upgrade run for prod/web with shum, show me its details,
and if it's in failed or rolled_back status, walk me through restoring the
backup artifact it recorded.
```
