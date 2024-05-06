package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	proxy "github.com/shogo82148/go-sql-proxy"
)

/*
	ref: https://songmu.jp/riji/entry/2021-02-03-go-sql-embed-comment.html
	this example add context information to sql query.
	it is useful for debugging especially when you use ORN.
*/

type endpointKey string

func main() {
	sql.Register("mysql-proxy", proxy.NewProxyContext(&mysql.MySQLDriver{}, &proxy.HooksContext{
		PreQuery: func(c context.Context, stmt *proxy.Stmt, args []driver.NamedValue) (interface{}, error) {
			stmt.QueryString = "/* " + c.Value(endpointKey("endpoint")).(string) + " */ " + stmt.QueryString
			return nil, nil
		},
	}))
	dsl := os.Getenv("DSL")
	db, err := sql.Open("mysql-proxy", dsl)
	ctx := context.WithValue(context.Background(), endpointKey("endpoint"), "/test")
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		log.Fatal(err)
	}
}
