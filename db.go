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

const categorySchema string = `
CREATE TABLE [dbo].[category] (
	[id] [int] IDENTITY(1,1) NOT NULL,
	[name] nvarchar(100) NOT NULL,
	CONSTRAINT [PK_category] PRIMARY KEY CLUSTERED 
	(
		[id] ASC
	)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
	) ON [PRIMARY]
`

const bucketSchema string = `
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
`

const bucketItemSchema string = `
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
  `

const templateSchema string = `
CREATE TABLE [dbo].[template] (
	  [id] [int] IDENTITY(1,1) NOT NULL,
	  [name] nvarchar(100) NOT NULL,
	 CONSTRAINT [PK_template] PRIMARY KEY CLUSTERED 
	(
		[id] ASC
	)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
	) ON [PRIMARY]
	`

const templateItemSchema string = `
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
	`

// go run db.go category.go bucket.go bucketItem.go errors.go
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

	_, err = sess.Exec(categorySchema)
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

	_, err = sess.Exec(bucketSchema)
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

	_, err = sess.Exec(bucketItemSchema)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	bucketItemCollection := sess.Collection("bucketitem")
	bucketItemCollection.Insert(BucketItem{
		Name:        "Initial Deposit",
		BucketID:    buckets[0].Id,
		Transaction: time.Date(1941, time.January, 5, 0, 0, 0, 0, time.UTC),
		Deposit:     3.99,
		Withdraw:    0.0,
	})

	res = bucketItemCollection.Find()
	// Query all results and fill the birthdays variable with them.
	var bucketItems []BucketItem

	err = res.All(&bucketItems)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}
	// Printing to stdout.
	for _, bucketItem := range bucketItems {
		fmt.Printf("%s has ID:%d.\n",
			bucketItem.Name,
			bucketItem.ID,
		)
	}

	_, err = sess.Exec(templateSchema)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	_, err = sess.Exec(templateItemSchema)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}
}
