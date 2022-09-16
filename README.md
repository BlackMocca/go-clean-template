# go-clean-template

## Prepare running application
### Installing proto
follow link `https://medium.com/@titlebhoomtawathplinsut/%E0%B8%A1%E0%B8%B2%E0%B8%97%E0%B8%B3-grpc-service-%E0%B8%94%E0%B9%89%E0%B8%A7%E0%B8%A2-go-%E0%B8%81%E0%B8%B1%E0%B8%99-866d7452f5dd`
### Installing database 
run with command
```script
docker-compose up -d psql_db
docker-compose up -d adminer
```
 
then open 127.0.0.1:8080 then input
-   PostgresQL
-   psql_db
-   postgres
-   example

Click ok and choose `database organization` and click `SQL command` and input `create extension "uuid-ossp"`

### Migrate table
Installing migration tool from makefile
-   on arch amd64
```script
make install-migration
```
-   on arch arm64
```script
make install-migration migrate_url="https://github.com/BlackMocca/migrate/releases/download/v5.0/migrate.linux-arm64"
```

`Migrate table` from this script 
```script
make app.migration.up db_url="$connectionstring_of_db" path=migrations/database/postgres/organization
```
example:
```
make app.migration.up db_url="postgres://postgres:example@127.0.0.1:5432/organization?sslmode=disable" path=migrations/database/postgres/organization
```

`Delete table` from this script 
```script
make app.migration.down db_url="$connectionstring_of_db" path=migrations/database/postgres/organization
```

`Seed Data`:
```
make app.migration.seed db_url="postgres://postgres:example@127.0.0.1:5432/organization?sslmode=disable" path=migrations/database/postgres/organization/master
```

note `if migrate fail you must delete table schema_migration or change version and set dirty is false`

### Running application
```script
docker-compose up app
```

-   Fetch All Organize
```script
curl -XGET http://127.0.0.1:3000/v1/organizes
```

-   Create Organize
```script
curl --location -XPOST 'http://127.0.0.1:3000/v1/organizes' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
    "alias_name":     "ทวจ.",
    "org_type":       "PUBLIC",
    "private_tel_no": "112341",
    "admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
    "admin_2":        "2f66d3c9-3549-66de-9639-0f6ff6c8d15a"
}'
```

```script
curl --location -XPOST 'http://127.0.0.1:3000/v1/organizes' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":           "นักลงทุนดอยในตำนาน",
    "alias_name":     "นักดอย",
    "org_type":       "PUBLIC",
    "private_tel_no": "112341",
    "admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d712",
    "admin_2":        "2f66d3c9-3549-66de-9639-0f6ff6c8d134"
}'
```