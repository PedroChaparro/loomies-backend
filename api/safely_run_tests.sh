# Run the `npm clear_loomies:all` command from the `/algorithms/database` directory to 
# clear the wild loomies collection before running the tests 
cd ../algorithms/database
npm run clear_loomies:all

# Go back to the api directory
cd ../../api

# Run the tests scripts
go test ./... -coverprofile=coverage.out
go tool cover --html=coverage.out