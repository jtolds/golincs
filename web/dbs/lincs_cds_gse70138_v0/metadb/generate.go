package metadb

//go:generate dbx -d sqlite3 metadb.dbx.yaml metadb.dbx.sql schema
//go:generate dbx -d sqlite3 metadb.dbx.yaml metadb.dbx.go code
//go:generate sed -i "s/^package db$/package metadb/" metadb.dbx.go
//go:generate goimports -w metadb.dbx.go
