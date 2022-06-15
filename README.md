## Web Development With Go
This project follows along with the Jonathan Calhound book "Web Development With Go"

 ### Database
 The database is provided by the postgres alpine image.

 Edit `database.env.TEMPLATE` to add your desired username, password, and databasename. Then rename this file to `database.env`

 Gorm will be used for database management and data mapping. The User table is AutoMigrated using the UserService.