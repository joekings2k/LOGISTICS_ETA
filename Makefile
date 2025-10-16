ifneq ("$(wildcard local.app.env)","")
    include local.app.env
    export $(shell sed 's/=.*//' local.app.env)
else ifneq ("$(wildcard app.env)","")
    include app.env
    export $(shell sed 's/=.*//' app.env)
endif

createdb: 
	docker exec -it postgres12 createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)

dropdb:
	docker exec -it postgres12 dropdb --username=$(DB_USER) $(DB_NAME)

migrateup:
	migrate -path db/migration -database "postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable" -verbose down
migratecreate:
	migrate create -ext sql -dir db/migration -seq init_schema

sqlc:
	sqlc generate

server:
	go run $(MAIN_GO)

mock:
	mockgen -package=mockdb -destination=db/mock/store.go --build_flags=--mod=mod github.com/joekings2k/logistics-eta/db/sqlc Store

.PHONY: createdb dropdb migrateup migratedown sqlc migratecreate server mock