# Recent activity is unavailable for ClickHouse views over timestamp sorted tables

## What happened

Recent activity is unavailable for a source backed by an ordinary ClickHouse
View, even when the underlying MergeTree table is ordered by the timestamp
column and the view preserves that column. Searching the source still works.

`InspectSourceActivity` requires the source's timestamp to lead its sorting key.
An ordinary View has no sorting key of its own, so the check rejects it.
The existing safeguard is intentional; this report proposes an explicit opt-in
for bounded activity queries rather than removing it for all sources.

## Expected behavior

An administrator should be able to opt specific regular views into Recent
activity when their underlying query plan is suitable. Other sources should
retain the leading-timestamp-sort-key safeguard.

## Steps to reproduce

Create these objects in a disposable ClickHouse database:

```sql
CREATE TABLE default.logs_base
(
    Timestamp DateTime64(9, 'UTC'),
    Message String
)
ENGINE = MergeTree
ORDER BY Timestamp;

CREATE VIEW default.logs_visible AS SELECT * FROM default.logs_base;
```

Configure a Logchef ClickHouse source for `default.logs_visible` with
`Timestamp` as its timestamp field. Search works, but Recent activity is
unavailable because the View metadata has no leading timestamp sorting key.

## Environment

Logchef v2.6.1, upstream commit `a0b1cf58a39b2fdede2cf427bd7c782bb90d511a`.
ClickHouse 26.8.36.3. The application already queries recent activity using a
24-hour window and excludes future timestamps.

## Proposed solution

The branch `fix/bounded-view-activity` adds a comma-separated deployment
allowlist, `LOGCHEF_BOUNDED_ACTIVITY_VIEWS`, with exact `database.view` matches.
An entry is accepted only for engine `View` with a configured timestamp field.
An empty allowlist keeps the existing behavior. A different database, a prefix
match, an unlisted view, or another engine is not admitted by this exception.

The patch also enforces non-overridable budgets in the shared ingestion-activity
query path: 3 seconds, 2 threads, 10 million rows read and 2 GiB read.
These budgets apply to ordinary activity queries as well as opted-in views;
maintainers may prefer a separate query option to scope the additional scan
limits to opted-in views. Scan overflow throws an error instead of returning
misleading partial statistics. Per-source settings cannot remove these limits.

This avoids trying to prove arbitrary SQL view performance by parsing its
definition. The operator explicitly chooses the views; a slow plan still fails
within a bounded budget. It does not grant extra ClickHouse privileges or
enable activity for all views by default.

## Regression coverage

Tests cover exact allowlist matching, default rejection, missing metadata,
engine restrictions, fixed scan/deadline limits and preservation of the caller's
settings map. Existing activity query/time-window tests remain in place.
`just check` passes natively without a local Docker environment.
