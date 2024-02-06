# service-catalog

service-catalog is a REST API that provides endpoints to access and use a service catalog.
It uses the Gin web framework, PostgreSQL as a database and JWT authentication.

## Schema

There are three tables:

### users

| column   | type         |
|----------|--------------|
| username | varchar(20)  |
| password | varchar(255) |

### services

| column      | type         |
|-------------|--------------|
| user_id     | int (FK)     |
| name        | varchar(255) |
| description | text         |

### versions

| column     | type        |
|------------|-------------|
| service_id | int (FK)    |
| version    | varchar(50) |
| changelog  | text        |

All tables also share the following columns:

| column     | type      |
|------------|-----------|
| id         | serial    |
| created_at | timestamp |
| updated_at | timestamp |

## Usage

* Populate the `.env` file. Refer to the `.env.sample` file for a list of possible env vars.
* Run the server and a Postgres instance:
  ```bash
  docker-compose up
  ```

To view API documentation, navigate to `/swagger/index.html`.

### Tests

```bash
make test
```

At the moment, there are tests only for read operations on Services along with Versions.

### Design decisions

* JWT over session auth.
* Use of an ORM over plain SQL queries, since this is a simple CRUD app and there are no complex queries.
* Seperate migrations folder for visibility and flexibility with optional auto migrations.
* PostgreSQL as a data store since the app fits in the relational model well.
* DB Normalization: a `versions` column is present in the `services` table, to avoid having to do a JOIN when
  we just want to show the version itself or the number of versions for the service.

### Scope for improvement

* No authorization support.
* No unit tests for the database.
* Lack of support for updating and deleting versions.
