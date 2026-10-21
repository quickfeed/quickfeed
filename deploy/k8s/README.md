# QuickFeed on Kubernetes: the database

This folder runs QuickFeed on the UiS Kubernetes cluster with a database that
**is not stored in the pod**. Data survives pod restarts, crashes, image
upgrades and node failures.

There are two steps, as Hein asked:

| | Step 1: SQLite on a volume | Step 2: PostgreSQL in the cluster |
|---|---|---|
| Folder | `base/` | `postgres/` (builds on `base/`) |
| Where the data is | `qf.db` on a persistent volume, mounted at `/data` | a PostgreSQL server (StatefulSet) with its own persistent volume |
| QuickFeed pods | exactly 1 | 1 now; more is possible later (see "Scaling") |
| Upgrade | short downtime (old pod stops, new pod starts) | short downtime now; rolling upgrades possible later |
| Backups | volume snapshots (manual) | automatic nightly `pg_dump` (CronJob), kept 14 days |
| Code changes | none | small, in `database/` (see "Code changes") |

```
deploy/k8s/
├── base/                       Step 1
│   ├── namespace.yaml
│   ├── data-pvc.yaml           volume for qf.db, Let's Encrypt certificates and logs
│   ├── deployment.yaml         QuickFeed: 1 replica, strategy Recreate, volume at /data
│   └── service.yaml
├── postgres/                   Step 2 (adds to base/)
│   ├── postgres-statefulset.yaml   PostgreSQL 17 with its own volume
│   ├── postgres-service.yaml       address postgres:5432, inside the cluster only
│   ├── quickfeed-use-postgres.yaml sets QUICKFEED_DATABASE_URL on QuickFeed
│   ├── backup-cronjob.yaml         nightly pg_dump
│   ├── backup-pvc.yaml             separate volume for the dumps
│   └── postgres.env.example        database name, user, password (copy to postgres.env)
└── jobs/
    └── migrate-sqlite-to-postgres.yaml  one-time copy of the data from step 1 to step 2
```

## Why the data survives

On the UiS cluster (Talos on OpenStack), a PersistentVolumeClaim (PVC) creates
an **OpenStack Cinder volume**. It is a network disk that exists outside the
pod and outside the node:

- **Pod restart / crash / new image:** the new pod mounts the same volume.
- **Node failure:** Kubernetes starts the pod on another node, and Cinder
  attaches the same volume there. (This can take a few minutes, because the
  volume must first be detached from the failed node.)
- **`kubectl delete pod`, `kubectl rollout restart`:** no data is lost.
- Data is lost only if the PVC itself is deleted (`kubectl delete pvc`, or
  deleting the namespace). Never do that without a backup.

## Before you start

1. `kubectl get nodes` works against the UiS cluster.
2. There is a StorageClass for Cinder:

   ```sh
   kubectl get storageclass
   ```

   One should be marked `(default)`. If none is, write its name into
   `storageClassName` in `base/data-pvc.yaml` and `postgres/postgres-statefulset.yaml`.
3. The QuickFeed image is built and pushed to GHCR by `.github/workflows/image.yml`.
   Set its tag (the commit SHA) in `base/kustomization.yaml` and `jobs/kustomization.yaml`:

   ```sh
   (cd deploy/k8s/base && kustomize edit set image quickfeed=ghcr.io/mahdavizeinab/quickfeed:<commit-sha>)
   ```

   If the GHCR package is private, the cluster also needs an image pull secret.
4. QuickFeed's own configuration (GitHub App, `DOMAIN`, secrets; see
   `.env-template`) is in a Secret named `quickfeed-env`:

   ```sh
   kubectl create namespace quickfeed
   kubectl -n quickfeed create secret generic quickfeed-env --from-env-file=.env
   ```

Exposing QuickFeed (Ingress / TLS passthrough, `qf-test.ux.uis.no`) is not
part of this folder; it uses the `quickfeed` Service on ports 443 and 80.

## Step 1: SQLite on a persistent volume

```sh
kubectl apply -k deploy/k8s/base
kubectl -n quickfeed get pvc,pods
```

The PVC should be `Bound` and the pod `Running`. QuickFeed creates an empty
database at `/data/qf.db` on first start.

To start with the anonymized test database instead (only for testing):

```sh
POD=$(kubectl -n quickfeed get pod -l app.kubernetes.io/name=quickfeed -o name)
kubectl -n quickfeed cp doc/dev/test.db ${POD#pod/}:/data/qf.db
kubectl -n quickfeed rollout restart deploy/quickfeed
```

**Backups in step 1:** take a snapshot of the Cinder volume. Stop QuickFeed
first so the file is not changed during the snapshot:

```sh
kubectl -n quickfeed scale deploy/quickfeed --replicas=0
openstack volume list            # find the volume of PVC quickfeed-data (pvc-...)
openstack volume snapshot create --volume <volume-id> --force quickfeed-$(date +%F)
kubectl -n quickfeed scale deploy/quickfeed --replicas=1
```

## Step 2: PostgreSQL in the cluster

```sh
cp deploy/k8s/postgres/postgres.env.example deploy/k8s/postgres/postgres.env
# edit postgres.env: set POSTGRES_PASSWORD (letters and digits only), e.g. openssl rand -hex 24
kubectl apply -k deploy/k8s/postgres
kubectl -n quickfeed get pods
```

QuickFeed waits until PostgreSQL is ready (init container `wait-for-postgres`),
then logs `Using PostgreSQL database from QUICKFEED_DATABASE_URL` and creates
its tables. The volume at `/data` stays, for the Let's Encrypt certificates
and the logs.

### Moving from SQLite to PostgreSQL (keeping the data)

If step 1 already has real data, copy it once:

```sh
# 1. Stop QuickFeed, so nothing writes to qf.db during the copy
kubectl -n quickfeed scale deploy/quickfeed --replicas=0

# 2. Start PostgreSQL (QuickFeed stays stopped)
kubectl apply -k deploy/k8s/postgres
kubectl -n quickfeed scale deploy/quickfeed --replicas=0
kubectl -n quickfeed rollout status statefulset/postgres

# 3. Copy the data
kubectl apply -k deploy/k8s/jobs
kubectl -n quickfeed logs -f job/migrate-sqlite-to-postgres
```

The job prints the number of rows in each table in both databases and fails
if any count differs. It refuses to copy into a database that already has
data. `qf.db` is mounted read-only and is not changed, so going back to step 1
is always possible.

```sh
# 4. Start QuickFeed on PostgreSQL
kubectl -n quickfeed scale deploy/quickfeed --replicas=1
```

Note: if QuickFeed was started on PostgreSQL before the copy, it has already
created (empty) tables, and the copy refuses to run if any of them has rows.

### Backups and restore

`backup-cronjob.yaml` runs `pg_dump` every night at 03:15 and writes
`quickfeed-<date>-<time>.dump` to the `postgres-backups` volume. Dumps older
than 14 days are deleted.

```sh
# Make a backup now (the log shows the file name and size)
kubectl -n quickfeed create job --from=cronjob/postgres-backup backup-now
kubectl -n quickfeed logs job/backup-now

# Copy a backup out of the cluster (keep a copy outside the cluster, too)
kubectl -n quickfeed exec -i postgres-0 -- sh -c 'pg_dump -U "$POSTGRES_USER" -Fc "$POSTGRES_DB"' > quickfeed.dump
```

Restore (replaces the current data; stop QuickFeed first):

```sh
kubectl -n quickfeed scale deploy/quickfeed --replicas=0
kubectl -n quickfeed exec -i postgres-0 -- sh -c \
  'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner' < quickfeed.dump
kubectl -n quickfeed scale deploy/quickfeed --replicas=1
```

The backup volume is a separate Cinder volume, so a backup survives if the
database volume is damaged or deleted. It does not survive if the whole
OpenStack project is lost; that is why a copy should also be kept outside.

## Checking that the data survives (for the demo / report)

```sh
# 1. Log in to QuickFeed and create something (e.g. a course), or check row counts:
kubectl -n quickfeed exec postgres-0 -- sh -c 'psql -U "$POSTGRES_USER" "$POSTGRES_DB" -c "select count(*) from users"'

# 2. Pod failure: delete the QuickFeed pod and the database pod
kubectl -n quickfeed delete pod -l app.kubernetes.io/name=quickfeed
kubectl -n quickfeed delete pod postgres-0
kubectl -n quickfeed get pods -w        # wait until both are Running again

# 3. Upgrade: deploy a new image tag, or
kubectl -n quickfeed rollout restart deploy/quickfeed

# 4. Node failure: move the database to another node
kubectl get pod -n quickfeed postgres-0 -o wide          # note the node
kubectl cordon <node> && kubectl -n quickfeed delete pod postgres-0
kubectl get pod -n quickfeed postgres-0 -o wide          # now on another node
kubectl uncordon <node>

# After each step, the data from step 1 is still there.
```

## Code changes (for step 2)

QuickFeed was written for SQLite. Hein expected switching to PostgreSQL to be
"mostly a one-line change". The driver change is small, but running QuickFeed's
full test suite against PostgreSQL showed several places where PostgreSQL is
stricter than SQLite:

| Problem found by the tests | Fix |
|---|---|
| Driver hard-coded to SQLite (`database/gormdb.go`) | `NewGormDB` chooses PostgreSQL when the source is a `postgres://` URL; `main.go` reads `QUICKFEED_DATABASE_URL` |
| Generated protobuf types use column type `datetime`, which PostgreSQL does not have | create `datetime` as a PostgreSQL domain over `timestamptz` (no change to generated code or to existing SQLite databases) |
| `group_id = 0` means "no group"; PostgreSQL enforces foreign keys, SQLite (as QuickFeed uses it) does not | do not create foreign key constraints on PostgreSQL, same behavior as before |
| `grades` has no primary key, and PostgreSQL needs a target for `ON CONFLICT DO UPDATE` | add the unique index (`submission_id, user_id`) as target |
| `ORDER BY 'order'` is a constant: SQLite ignores it, PostgreSQL rejects it | quote the column correctly (assignments are now really sorted by order) |
| Duplicate-group check compared SQLite's error text | `isUniqueViolation` recognizes both databases |
| Group update inserted each grade twice in one statement | fixed in `syncGroupGrades` |
| Rows inserted with a chosen ID: PostgreSQL's ID counter does not notice them | move the counter after such inserts |
| Go timestamps have nanoseconds, PostgreSQL stores microseconds | `database.Now()` rounds to microseconds |

The PostgreSQL-specific code is in `database/postgres.go`. SQLite still works
exactly as before and remains the default.

### Running the tests against PostgreSQL

The tests use a temporary SQLite file by default. With
`QUICKFEED_TEST_DATABASE_URL` set, every test gets its own new PostgreSQL
database instead:

```sh
docker run -d --name qf-pg -e POSTGRES_PASSWORD=test -p 5432:5432 postgres:17.6
QUICKFEED_TEST_DATABASE_URL='postgres://postgres:test@localhost:5432/postgres?sslmode=disable' \
  go test ./...
```

Results: all 21 test packages pass on SQLite and on PostgreSQL (tests that
need Docker or a GitHub token are skipped, as in any local run without them).
`TestCopyTo` copies the anonymized test database (`doc/dev/test.db`) and
checks every row of every table after the copy.

## Scaling (more than one QuickFeed pod)

PostgreSQL removes the database as the reason for running only one pod: many
pods can use one PostgreSQL server, but not one SQLite file. For now QuickFeed
still runs as one pod with `strategy: Recreate`, because:

- the `/data` volume (certificates, logs) is ReadWriteOnce;
- QuickFeed keeps some state in memory and runs the student tests itself
  (Docker), which would need to be shared or moved before running several
  copies.

These are not database questions and are left for later.
