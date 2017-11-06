package main

import (
	"fmt"
	"log"

	db "upper.io/db.v3"
	"upper.io/db.v3/mssql"
)

var settings = mssql.ConnectionURL{
	Host:     "127.0.0.1", // MSSQL server IP or name.
	Database: "budget2",   // Database name.
	User:     "budgetUser",
	Password: "budgetPassword",
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
	`

// go run db.go category.go bucket.go bucketItem.go errors.go template.go templateItem.go
func dbInitMain() {
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

	_, err = sess.Exec(templateSchema)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}

	templateCollection := sess.Collection("template")
	templateCollection.Insert(Template{
		Name: "Bimonthly paycheck",
	})
	res = templateCollection.Find()
	// Query all results and fill the birthdays variable with them.
	var templates []Template

	err = res.All(&templates)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}
	// Printing to stdout.
	for _, template := range templates {
		fmt.Printf("%s has ID:%d.\n",
			template.Name,
			template.Id,
		)
	}

	_, err = sess.Exec(templateItemSchema)
	if err != nil {
		fmt.Printf("Table already created %q\n", err)
	}
	templateItemCollection := sess.Collection("templateitem")
	templateItemCollection.Insert(TemplateItem{
		Name:       "Deposit",
		BucketID:   buckets[0].Id,
		TemplateID: templates[0].Id,
		Deposit:    2.99,
		Withdraw:   1.45,
	})
	res = templateItemCollection.Find()
	// Query all results and fill the birthdays variable with them.
	var templateItems []TemplateItem

	err = res.All(&templateItems)
	if err != nil {
		log.Fatalf("res.All(): %q\n", err)
	}
	// Printing to stdout.
	for _, templateItem := range templateItems {
		fmt.Printf("%s has ID:%d.\n",
			templateItem.Name,
			templateItem.ID,
		)
	}
}

func dbNewBucketItem(bucketItem *BucketItem) error {
	sess, err := mssql.Open(settings)
	if err != nil {
		return err
	}
	defer sess.Close()

	bucketItemCollection := sess.Collection("bucketitem")
	bucketItemCollection.Insert(bucketItem)
	res := bucketItemCollection.Find()
	err = res.One(bucketItem)

	return err
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
	bucketCollection.Insert(bucket)
	res := bucketCollection.Find()
	err = res.One(bucket)

	return err
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
	categoryCollection.Insert(category)
	res := categoryCollection.Find()
	err = res.One(category)

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
	templateCollection.Insert(template)
	res := templateCollection.Find()
	err = res.One(template)

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
	templateItemCollection.Insert(templateItem)
	res := templateItemCollection.Find()
	err = res.One(templateItem)

	return err
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
