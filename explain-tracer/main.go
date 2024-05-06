package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	proxy "github.com/shogo82148/go-sql-proxy"
)

/*
	this example get explain result of running query.
*/

type endpointKey string

func main() {
	sql.Register("mysql-proxy", proxy.NewProxyContext(&mysql.MySQLDriver{}, &proxy.HooksContext{
		PreQuery: func(c context.Context, stmt *proxy.Stmt, args []driver.NamedValue) (interface{}, error) {
			expQuery := "EXPLAIN " + stmt.QueryString
			if queryerCtx, ok := stmt.Conn.Conn.(driver.QueryerContext); ok {
				rows, err := queryerCtx.QueryContext(c, expQuery, args)
				if err != nil {
					log.Fatalf("Explain failed: %v", err)
				}
				defer rows.Close()
				columns := rows.Columns()
				values := make([]driver.Value, len(columns))
				for {
					err := rows.Next(values)
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Fatalf("Explain failed: %v", err)
					}
					for i, value := range values {
						log.Printf("%s: %s", columns[i], value)
					}
				}
			}
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

	rows, err := db.QueryContext(ctx, "SELECT * from users where id = 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("id: %d, name: %s", id, name)
	}
}
