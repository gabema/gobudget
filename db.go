package main

import (
	"fmt"
	"log"
	"time"

	"upper.io/db.v3/mssql"
)

var settings = mssql.ConnectionURL{
	Host:     "127.0.0.1", // MSSQL server IP or name.
	Database: "budget2",   // Database name.
	User:     "budgetUser",
	Password: "budgetPassword",
}

type Birthday struct {
	// Name maps the "Name" property to the "name" column
	// of the "birthday" table.
	Name string `db:"name"`

	// Born maps the "Born" property to the "born" column
	// of the "birthday" table.
	Born time.Time `db:"born"`
}

type Category struct {
	// Name maps the "Name" property to the "name" column
	// of the "birthday" table.
	Name string `db:"name"`

	// Id of the category
	Id int `db:"id,omitempty"`
}

func main() {
	// Attemping to establish a connection to the database.
	sess, err := mssql.Open(settings)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}
	defer sess.Close() // Remember to close the database session.

	sess.Collection("category").Truncate()
	sess.Collection("template").Truncate()
	sess.Collection("templateitem").Truncate()
	sess.Collection("bucket").Truncate()
	sess.Collection("bucketitem").Truncate()

	// _, err = sess.Exec(`
	// 	CREATE TABLE [dbo].[birthday](
	// 		[id] [int] IDENTITY(1,1) NOT NULL,
	// 		[born] [datetime2](0) NOT NULL,
	// 		[name] [varchar](50) NOT NULL,
	// 	 CONSTRAINT [PK_birthday] PRIMARY KEY CLUSTERED
	// 	(
	// 		[id] ASC
	// 	)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
	// 	) ON [PRIMARY]
	// 	`)
	// if err != nil {
	// 	fmt.Printf("Table already created %q\n", err)
	// }
	_, err = sess.Exec(`
		CREATE TABLE [dbo].[category] (
			[id] [int] IDENTITY(1,1) NOT NULL,
			[name] nvarchar(100) NOT NULL,
		   CONSTRAINT [PK_category] PRIMARY KEY CLUSTERED 
		  (
			  [id] ASC
		  )WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
		  ) ON [PRIMARY]
		`)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}
	categoryCollection := sess.Collection("category")
	categoryCollection.Insert(Category{
		Name: "house",
	})
	categoryCollection.Insert(Category{
		Name: "living",
	})
	res := categoryCollection.Find()
	// Query all results and fill the birthdays variable with them.
	var categories []Category

	err = res.All(&categories)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}

	// Printing to stdout.
	for _, category := range categories {
		fmt.Printf("%s has ID:%d.\n",
			category.Name,
			category.Id,
		)
	}

	_, err = sess.Exec(`
		CREATE TABLE [dbo].[bucket] (
			[id] [int] IDENTITY(1,1) NOT NULL,
			[categoryID] [int] NOT NULL,
			[name] nvarchar(100) NOT NULL,
			[description] nvarchar(1000) NOT NULL DEFAULT N'',
			[isLiquid] bit NOT NULL DEFAULT 1
		   CONSTRAINT [PK_bucket] PRIMARY KEY CLUSTERED 
		  (
			  [id] ASC
		  )WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
		  ) ON [PRIMARY]
		`)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}
	bucketCollection := sess.Collection("bucket")
	bucketCollection.Insert(Bucket{
		Name:       "Gas",
		CategoryID: categories[0].Id,
		IsLiquid:   true,
	})
	bucketCollection.Insert(Bucket{
		Name:       "Gabe's Personal",
		CategoryID: categories[1].Id,
		IsLiquid:   false,
	})
	res = bucketCollection.Find()
	// Query all results and fill the birthdays variable with them.
	var buckets []Bucket

	err = res.All(&buckets)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}
	// Printing to stdout.
	for _, bucket := range buckets {
		fmt.Printf("%s has ID:%d.\n",
			bucket.Name,
			bucket.Id,
		)
	}

	_, err = sess.Exec(`
		CREATE TABLE [dbo].[bucketitem] (
			[id] [int] IDENTITY(1,1) NOT NULL,
			[bucketID] [int] NOT NULL,
			[transaction] datetime2(0) NOT NULL,
			[name] nvarchar(100) NOT NULL,
			[deposit] decimal(10,2) NOT NULL DEFAULT 0.00,
			[withdrawl] decimal(10,2) NOT NULL DEFAULT 0.00
		   CONSTRAINT [PK_bucketitem] PRIMARY KEY CLUSTERED 
		  (
			  [id] ASC
		  )WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
		  ) ON [PRIMARY]
		  `)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	_, err = sess.Exec(`		
		CREATE TABLE [dbo].[template] (
			[id] [int] IDENTITY(1,1) NOT NULL,
			[name] nvarchar(100) NOT NULL,
		   CONSTRAINT [PK_template] PRIMARY KEY CLUSTERED 
		  (
			  [id] ASC
		  )WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
		  ) ON [PRIMARY]
		  `)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	_, err = sess.Exec(`
		CREATE TABLE [dbo].[templateitem] (
			[id] [int] IDENTITY(1,1) NOT NULL,
			[templateID int NOT NULL,
			[bucketID] int NOT NULL,
			[name] nvarchar(100) NOT NULL,
			[deposit] decimal(10,2) NOT NULL DEFAULT 0.00,
			[withdraw] decimal(10,2) NOT NULL DEFAULT 0.00
		   CONSTRAINT [PK_templateitem] PRIMARY KEY CLUSTERED 
		  (
			  [id] ASC
		  )WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
		  ) ON [PRIMARY]
		  `)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	// Pointing to the "birthday" table.
	birthdayCollection := sess.Collection("birthday")

	// Attempt to remove existing rows (if any).
	err = birthdayCollection.Truncate()
	if err != nil {
		log.Fatalf("Truncate(): %q\n", err)
	}

	// Inserting some rows into the "birthday" table.
	birthdayCollection.Insert(Birthday{
		Name: "Hayao Miyazaki",
		Born: time.Date(1941, time.January, 5, 0, 0, 0, 0, time.UTC),
	})

	birthdayCollection.Insert(Birthday{
		Name: "Nobuo Uematsu",
		Born: time.Date(1959, time.March, 21, 0, 0, 0, 0, time.UTC),
	})

	birthdayCollection.Insert(Birthday{
		Name: "Hironobu Sakaguchi",
		Born: time.Date(1962, time.November, 25, 0, 0, 0, 0, time.UTC),
	})

	// Let's query for the results we've just inserted.
	res = birthdayCollection.Find()

	// Query all results and fill the birthdays variable with them.
	var birthdays []Birthday

	err = res.All(&birthdays)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}

	// Printing to stdout.
	for _, birthday := range birthdays {
		fmt.Printf("%s was born in %s.\n",
			birthday.Name,
			birthday.Born.Format("January 2, 2006"),
		)
	}
}
