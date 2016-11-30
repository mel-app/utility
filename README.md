# MEL app backend CLI utility #

This utility provides several functions that needed either increased
permissions or have yet to be implemented through the API.

To use, set the DATABASE\_TYPE and DATABASE\_URL environmental variables and
run the utility.

DATABASE\_TYPE should probably be set to postgres if connecting to the heroku
database, while DATABASE\_URL can be found under the "Database Credentials"
section in data.heroku.com (listed as "URI").

Running the utility with no arguments prints a short description of the
available tools and how to use them.

