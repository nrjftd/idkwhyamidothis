package models

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/uptrace/bun"
)

var constraints = map[string]map[string]string{
	"users": {
		"email_format": `ALTER TABLE users ADD CONSTRAINT email_format CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');`,
		"first_name_check": `ALTER TABLE users
		ADD CONSTRAINT first_name_length CHECK (LENGTH(first_name) BETWEEN 2 AND 100);`,
		"last_name_check": `ALTER TABLE users
		ADD CONSTRAINT last_name_length CHECK (LENGTH(last_name) BETWEEN 2 AND 100);`,
		"password_check": `ALTER TABLE users
		ADD CONSTRAINT password_length CHECK(LENGTH(password) >=6);`,
		"phone_check": `ALTER TABLE users
	ADD CONSTRAINT phone_format CHECK (phone ~ '^[0-9]{10}$;`,
		"user_type_check": `ALTER TABLE users
	ADD CONSTRAINT check_user_type CHECK (user_type IN ('ADMIN', 'USER'));`,
	},

	"task": {
		"description_check": `ALTER TABLE task 
	ADD CONSTRAINT description_length CHECK (LENGTH(description) >=10);`,
		"title_check": `ALTER TABLE task
	ADD CONSTRAINT title_length CHECK (LENGTH(title) BETWEEN 2 AND 100;`,
	},
}

func CreateTable(db *bun.DB) {
	ctx := context.Background()
	_, err := db.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	// //email
	// _, err = db.ExecContext(ctx, ` ALTER TABLE users
	// ADD
	//  email_format

	// return existingConstraints, nil CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');`)
	// if err != nil {
	// 	log.Fatalf("Failed to add constraint: %v", err)
	// }
	// //
	// //first name
	// _, err = db.ExecContext(ctx, `ALTER TABLE users
	// 	ADD CONSTRAINT first_name_length CHECK (LENGTH(first_name) BETWEEN 2 AND 100);
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add first name constraint: %v", err)
	// }
	// //last name
	// _, err = db.ExecContext(ctx, `ALTER TABLE users
	// 	ADD CONSTRAINT last_name_length CHECK (LENGTH(last_name) BETWEEN 2 AND 100);
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add last name constraint: %v", err)
	// }
	// //password
	// _, err = db.ExecContext(ctx, `ALTER TABLE users
	// 	ADD CONSTRAINT password_length CHECK(LENGTH(password) >=6);
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add password constraint: %v", err)
	// }
	// //phone
	// _, err = db.ExecContext(ctx, `ALTER TABLE users
	// ADD CONSTRAINT phone_format CHECK (phone ~ '^[0-9]{10}$');
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add phone constraint: %v", err)
	// }

	// //user type
	// _, err = db.ExecContext(ctx, `ALTER TABLE users
	// ADD CONSTRAINT check_user_type CHECK (user_type IN ('ADMIN', 'USER'));
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add phone constraint: %v", err)
	// }

	_, err = db.NewCreateTable().Model((*Task)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to create tasks table: %v", err)
	}
	// //description
	// _, err = db.ExecContext(ctx, `ALTER TABLE task
	// ADD CONSTRAINT description_length CHECK (LENGTH(description) >=10);`)
	// if err != nil {
	// 	log.Fatalf("Failed to add description constraint: %v", err)
	// }

	// //title
	// _, err = db.ExecContext(ctx, `ALTER TABLE task
	// ADD CONSTRAINT title_length CHECK (LENGTH(title) BETWEEN 2 AND 100);
	// `)
	// if err != nil {
	// 	log.Fatalf("Failed to add title constraint: %v", err)
	// }

	_, err = db.NewCreateTable().Model((*RefreshToken)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to create refresh_tokens table: %v", err)
	}
	log.Println("All table created successfully")
}

func GetExistingConstraint(db *bun.DB, tables []string) (map[string]string, error) {
	ctx := context.Background()
	tableName := fmt.Sprintf("('%s')", strings.Join(tables, "', '"))
	query := fmt.Sprintf(`
		SELECT c.relname AS table_name,
				string_agg(con.conname, ', ') AS constraints
		FROM pg_constraint con
		JOIN pg_class c ON con.conrelid = c.oid
		JOIN pg_namespace n ON c.conrelid = n.oid 
		WHERE n.nspname ='public' AND c.relname IN %s
		GROUP BY c.relname;
	`, tableName)
	existingConstraints := make(map[string]string)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tableName, constraints string
		if err := rows.Scan(&tableName, &constraints); err != nil {
			return nil, err
		}
		existingConstraints[tableName] = constraints

	}
	return existingConstraints, nil
}

func AddConstraints(db *bun.DB) {
	var tables []string
	for table := range constraints {
		tables = append(tables, table)
	}
	existingConstraint, err := GetExistingConstraint(db, tables)
	if err != nil {
		log.Fatalf("Failed to fetch constraint: %v", err)
	}
	var queries []string
	for table, constraintList := range constraints {
		existing := existingConstraint[table]
		for name, query := range constraintList {
			if !strings.Contains(existing, name) {
				queries = append(queries, query)
			}
		}
	}
	if len(queries) > 0 {
		ctx := context.Background()
		_, err := db.ExecContext(ctx, strings.Join(queries, " "))
		if err != nil {
			log.Fatalf("Failed to add constraint: %v", err)
		}
		log.Println("constraints added successfully: ", queries)

	} else {
		log.Println("All constraint already exist, skip")
	}
}
