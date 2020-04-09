# Billboard_Hot.100
### Scrapes Billboard.com for Hot 100 chart data, from Aug 1958 to the present.

#### Develop:

    go run main.go

#### Build:

    go build

#### Usage:

    Billboard_Hot.100 startDate [endDate]

Dates must be in the form of: YYYY-MM-DD  
Up to twenty weeks at a time may be downloaded per execution.  
endDate defaults to twenty weeks past the startDate.

Results are JSON files that are put into a newly created 'data' folder, and then in folders by year.

The Hot 100 list is released each Saturday for the upcoming week, starting on Sunday. The official listed date is that of the Monday of that week.
