# Billboard_Hot.100
### Scrapes Billboard.com for Hot 100 chart data, from Aug 1958 to the present.

#### Develop and run:

    go run main.go

#### Build executable:

    go build

#### Usage:

    Billboard_Hot.100 [option] startDate [endDate]

Dates must be in the form of: YYYY-MM-DD  
endDate defaults to twenty weeks past the startDate.  
By default JSON files are created, but it will output CSV files if the -csv flag is specified.  
If a data file already exists, that week will be skipped. Delete it to retrieve that data again.  

Options:  
-csv  Writes CSV data files instead of JSON  
-verify  Verifies that all JSON and CSV data files in the data directory each contain 100 songs. (When the verify option is used, the startDate and endDate are ignored.) Invalid data files are deleted.  

Results are JSON or CSV files that are put into a newly created 'data' folder, and then in folders by year.  

Billboard releases the Hot 100 list each Saturday for the upcoming week, starting on Sunday. The official listed date is that of the Monday of that week.  
