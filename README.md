# keyValueStore-golang

## Implement a Key-Value Store
In this problem, you'll implement an in-memory key/value store service similar to Redis. 
For simplicity's sake, instead of dealing with IO you only need to implement this 
without persistance / values are lost if the service is terminated.


## Data Commands
The database accepts the following commands to operate on keys:

* `SET name value` – Set the variable `name` to the value `value`. For
  simplicity `value` may be an integer.
* `GET name` – Print out the value of the variable `name`, or `NULL` if that
  variable is not set.
* `UNSET name` – Unset the variable `name`, making it just like that variable
  was never set.


+ Here, ">" is the prompt sign for the user to provide input for the application.


```
INPUT	            OUTPUT
--------------------------
> SET a 10
> GET a
10
> UNSET a
> GET a
NULL
>


INPUT	            OUTPUT
--------------------------
> SET b 10
> SET b 30
> GET b
30
>
```

## Transaction Commands
In addition to the above data commands, your program should also support
database transactions by also implementing these commands:

* `BEGIN` – Open a new transaction block. **Transactions can be nested;** a
  `BEGIN` can be issued inside of an existing block.
* `ROLLBACK` – Undo commands issued in the current transaction, and closes it.
  Returns an error if no transaction is in progress.
* `COMMIT` – Close **all** open transactions, permanently applying the changes
  made in them. Returns an error if no transaction is in progress.

Any data command that is run outside of a transaction should commit
immediately. Here are some example command sequences:


```

INPUT	          OUTPUT
------------------------
> BEGIN
> SET a 30
> BEGIN
> SET a 40
> COMMIT
> GET a
40
> ROLLBACK
NO TRANSACTION
> END

INPUT	          OUTPUT
------------------------
> BEGIN
> SET a 10
> GET a
10
> BEGIN
> SET a 20
> GET a
20
> ROLLBACK
> GET a
10
> ROLLBACK
> GET a
NULL
> END

```



