# random things made on Praktyki JABA DABA >:3

# Initialise a module (once)
go mod init mycsvparser

# Build/run
go run . -dir ./my-csv-folder -op "<" -price 25.00   # products cheaper than $25
go run . -dir ./my-csv-folder -op "=" -price 49.99   # exactly $49.99
go run . -dir ./my-csv-folder -op ">" -price 100    # more expensive than $100
# JSON output:
go run . -dir ./my-csv-folder -op ">" -price 100 -json

-dir	Path to the folder that holds all the CSV files you want to scan.
-op	Comparison operator for the price filter: < (lower), = (equal), > (higher).
-price	Numeric threshold used together with -op.
-json	Optional; prints the result as pretty‑printed JSON instead of a plain table.
