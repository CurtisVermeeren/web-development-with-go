## Web Development With Go
This project follows along with the Jonathan Calhound book "Web Development With Go"

### Project setup
Edit `.env.TEMPLATE` to add all required parameters. Then rename this file to `.env`

Setup the database as described in the database section below.

### Database
The database is provided by the postgres alpine image.

Edit `database.env.TEMPLATE` to add your desired username, password, and databasename. Then rename this file to `database.env`

Gorm will be used for database management and data mapping. The User table is AutoMigrated using the UserService.

`userService.DestructiveReset()` can be called from the main method to reset the database for development.

### Images
Images uploaded to a gallery are stored in the server filesystem. The `images/` directory contains ids of all galleriers and their images. 

BOOKMARK --- 627