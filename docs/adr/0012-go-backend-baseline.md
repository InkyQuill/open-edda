# Go Backend Baseline

Writer's first backend baseline is Go with `chi` for HTTP routing, `sqlc` for database access, `goose` for migrations, and SQLite as the first database. SQLite fits the single-author self-hosted workload if the implementation uses WAL mode, keeps transactions short, avoids long-running agent work inside transactions, and treats Postgres as a later option if collaboration or heavier concurrency demands it.
