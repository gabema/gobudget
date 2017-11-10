package main

import (
	"fmt"
	"log"
	"time"

	db "upper.io/db.v3"
	"upper.io/db.v3/mssql"
)

var settings = mssql.ConnectionURL{
	Host:     readEnvOrDefault("DB_HOST_NAME", "127.0.0.1"), // MSSQL server IP or name.
	Database: readEnvOrDefault("DB_NAME", "budget2"),        // Database name.
	User:     readEnvOrDefault("DB_USER", "budgetUser"),
	Password: readEnvOrDefault("DB_PASSWORD", "budgetPassword"),
}

func dbInit() {
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

	categoryCollection := sess.Collection("category")
	categoryCollection.Insert(Category{
		Name: "house",
	})
	categoryCollection.Insert(Category{
		Name: "living",
	})

	bucketCollection := sess.Collection("bucket")
	bucketCollection.Insert(Bucket{
		Name:       "Gas",
		CategoryID: 1,
		IsLiquid:   true,
	})
	bucketCollection.Insert(Bucket{
		Name:       "Gabe's Personal",
		CategoryID: 1,
		IsLiquid:   false,
	})

	bucketItemCollection := sess.Collection("bucketitem")
	bucketItemCollection.Insert(BucketItem{
		Name:        "Initial Deposit",
		BucketID:    1,
		Deposit:     1.99,
		Withdraw:    0.44,
		Transaction: time.Now(),
	})
	bucketCollection.Insert(Bucket{
		Name:       "Gabe's Personal",
		CategoryID: 1,
		IsLiquid:   false,
	})

	templateCollection := sess.Collection("template")
	templateCollection.Insert(Template{
		Name: "Bimonthly paycheck",
	})

	templateItemCollection := sess.Collection("templateitem")
	templateItemCollection.Insert(TemplateItem{
		Name:       "Deposit",
		BucketID:   1,
		TemplateID: 1,
		Deposit:    2.99,
		Withdraw:   1.45,
	})
}

func dbDropTables() error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close() // Remember to close the database session.

	if _, err = sess.Exec("drop TABLE [dbo].[category];"); err != nil {
		fmt.Printf("Err: %q\n", err)
	}

	if _, err = sess.Exec("drop TABLE [dbo].[bucket];"); err != nil {
		fmt.Printf("Err: %q\n", err)
	}

	if _, err = sess.Exec("drop TABLE [dbo].[bucketitem];"); err != nil {
		fmt.Printf("Err: %q\n", err)
	}

	if _, err = sess.Exec("drop TABLE [dbo].[template];"); err != nil {
		fmt.Printf("Err: %q\n", err)
	}

	if _, err = sess.Exec("drop TABLE [dbo].[templateitem];"); err != nil {
		fmt.Printf("Err: %q\n", err)
	}

	return err
}

func dbCreateTables() error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close() // Remember to close the database session.

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
		  ) ON [PRIMARY],
		  CONSTRAINT FK_bucket_category FOREIGN KEY (categoryID)     
		  REFERENCES dbo.category ([id])
		`)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
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
		  ) ON [PRIMARY],
		  CONSTRAINT FK_bucketitem_bucket FOREIGN KEY (bucketID)     
		  REFERENCES dbo.bucket ([id])
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
			[templateID] int NOT NULL,
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

	return err
}

func dbNewBucketItem(bucketItem *BucketItem) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketItemCollection := sess.Collection("bucketitem")
	return bucketItemCollection.InsertReturning(bucketItem)
}

func dbGetBucketItems() ([]*BucketItem, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var bucketItems []*BucketItem
	bucketItemCollection := sess.Collection("bucketitem")
	res := bucketItemCollection.Find()
	err = res.All(&bucketItems)

	return bucketItems, err
}

func dbGetBucketItem(id int) (*BucketItem, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var bucketItem BucketItem
	bucketItemCollection := sess.Collection("bucketitem")
	res := bucketItemCollection.Find(db.Cond{"id": id})
	err = res.One(&bucketItem)

	return &bucketItem, err
}

func dbUpdateBucketItem(id int, bucketItem *BucketItem) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketItemCollection := sess.Collection("bucketitem")
	res := bucketItemCollection.Find(db.Cond{"id": id})
	err = res.Update(bucketItem)
	if err != nil {
		return err
	}
	err = res.One(bucketItem)

	return err
}

func dbRemoveBucketItem(id int) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketItemCollection := sess.Collection("bucketitem")
	res := bucketItemCollection.Find(db.Cond{"id": id})
	err = res.Delete()

	return err
}

func dbNewBucket(bucket *Bucket) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketCollection := sess.Collection("bucket")
	return bucketCollection.InsertReturning(bucket)
}

func dbGetBuckets() ([]*Bucket, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var buckets []*Bucket
	bucketCollection := sess.Collection("bucket")
	res := bucketCollection.Find()
	err = res.All(&buckets)

	return buckets, err
}

func dbGetBucket(id int) (*Bucket, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var bucket Bucket
	bucketCollection := sess.Collection("bucket")
	res := bucketCollection.Find(db.Cond{"id": id})
	err = res.One(&bucket)

	return &bucket, err
}

func dbUpdateBucket(id int, bucket *Bucket) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketCollection := sess.Collection("bucket")
	res := bucketCollection.Find(db.Cond{"id": id})
	err = res.Update(bucket)
	if err != nil {
		return err
	}
	err = res.One(bucket)

	return err
}

func dbRemoveBucket(id int) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketCollection := sess.Collection("bucket")
	res := bucketCollection.Find(db.Cond{"id": id})
	err = res.Delete()

	return err
}

func dbNewCategory(category *Category) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	categoryCollection := sess.Collection("category")
	err = categoryCollection.InsertReturning(category)
	return err
}

func dbGetCategories() ([]*Category, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var categories []*Category
	categoryCollection := sess.Collection("category")
	res := categoryCollection.Find()
	err = res.All(&categories)

	return categories, err
}

func dbGetCategory(id int) (*Category, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var category Category
	categoryCollection := sess.Collection("category")
	res := categoryCollection.Find(db.Cond{"id": id})
	err = res.One(&category)

	return &category, err
}

func dbUpdateCategory(id int, category *Category) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	categoryCollection := sess.Collection("category")
	res := categoryCollection.Find(db.Cond{"id": id})
	err = res.Update(category)
	if err != nil {
		return err
	}
	err = res.One(category)

	return err
}

func dbRemoveCategory(id int) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	categoryCollection := sess.Collection("category")
	res := categoryCollection.Find(db.Cond{"id": id})
	err = res.Delete()

	return err
}

func dbNewTemplate(template *Template) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateCollection := sess.Collection("template")
	err = templateCollection.InsertReturning(template)

	return err
}

func dbGetTemplates() ([]*Template, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var templates []*Template
	templateCollection := sess.Collection("template")
	res := templateCollection.Find()
	err = res.All(&templates)

	return templates, err
}

func dbGetTemplate(id int) (*Template, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var template Template
	templateCollection := sess.Collection("template")
	res := templateCollection.Find(db.Cond{"id": id})
	err = res.One(&template)

	return &template, err
}

func dbUpdateTemplate(id int, template *Template) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateCollection := sess.Collection("template")
	res := templateCollection.Find(db.Cond{"id": id})
	err = res.Update(template)
	if err != nil {
		return err
	}
	err = res.One(template)

	return err
}

func dbRemoveTemplate(id int) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateCollection := sess.Collection("template")
	res := templateCollection.Find(db.Cond{"id": id})
	err = res.Delete()

	return err
}

func dbNewTemplateItem(templateItem *TemplateItem) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateItemCollection := sess.Collection("templateitem")
	return templateItemCollection.InsertReturning(templateItem)
}

func dbGetTemplateItems() ([]*TemplateItem, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var templateItems []*TemplateItem
	templateItemCollection := sess.Collection("templateitem")
	res := templateItemCollection.Find()
	err = res.All(&templateItems)

	return templateItems, err
}

func dbGetTemplateItem(id int) (*TemplateItem, error) {
	sess, err := mssql.Open(settings)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var templateItem TemplateItem
	templateItemCollection := sess.Collection("templateitem")
	res := templateItemCollection.Find(db.Cond{"id": id})
	err = res.One(&templateItem)

	return &templateItem, err
}

func dbUpdateTemplateItem(id int, templateItem *TemplateItem) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateItemCollection := sess.Collection("templateitem")
	res := templateItemCollection.Find(db.Cond{"id": id})
	err = res.Update(templateItem)
	if err != nil {
		return err
	}
	err = res.One(templateItem)

	return err
}

func dbRemoveTemplateItem(id int) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	templateItemCollection := sess.Collection("templateitem")
	res := templateItemCollection.Find(db.Cond{"id": id})
	err = res.Delete()

	return err
}
