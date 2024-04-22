# Leagueify API

Leagueify is an open source platform for managing sporting leagues of any size.

This server is written in [Go][go-website] using the [Echo][echo-website] framework.

## Getting Started

By default the API documentation can be found at http://localhost/api

Leagueify API uses easy to use Makefile commands to get up and running quickly. Currently, Leagueify API requires Go 1.22.0 in order to run, please install Go before proceeding with the following commands:

``` bash
# Install Dependencies
make init

# Run Leagueify Locally
make dev-start

# Stop Leagueify and Remove Docker Image
make dev-clean
```

**NOTE:** Until the email configuration has been developed, it would be best to add `"id":     account.ID,` to `internal/api/accounts:96`

## Contribution Requirements

Leagueify API makes use of automated checks to verify code quality. To ensure code quality, please run the following commands before creating a PR:

```bash
# Vet Code for Errors
make vet

# Format Code 
make format

# Clean Go Dependencies
make clean
```

[go-website]: https://go.dev
[echo-website]: https://echo.labstack.com
