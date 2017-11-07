# gobudget

A simple budget api web service written in GO and intended for Azure deployment and local development.

## Environment Settings

### HTTP_PLATFORM_PORT
Tells the Go server which http port to listen on

### DB_HOST_NAME
SQL Server Hostname: Default to "127.0.0.1"

### DB_NAME
The Name of the SQL Server database to use. Defaults to "budget2"

### DB_USER
The Username to connect to the SQL Server. Defaults to budgetUser

### DB_PASSWORD
The Password to connect to the SQL Server. Defaults to budgetPassword

## Deployment artifacts
The only files necessary to upload to the wwwroot directory is the go compiled executable and the web.config.