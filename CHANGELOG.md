# Change History

## April 1 2020: v1.0.1

  * **Improvements**

  	- Adds tests against sqlite3 to make sure the results are correct.

  * **Fixes**

    - Fixes a an issue where sometimes more fields than specified would be returned.


## April 1 2020: v1.0.0

  Major Breaking Change.

  * **Changes**

  	- Simplifies the payload definition of the lua function. Read the README.md for more info.

  * **Fixes**

    - Fixes a race condition triggered in the Go client's Lua library.

