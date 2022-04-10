# Fitness Scheduler

An assessment project for a Backend Engineer role.

## Running the project locally

### Prerequisites

- .env file with DSN and API_URL variables (Clark will send via email as needed)
- install [Air](https://github.com/cosmtrek/air#installation) if live reload is desired

### Running with live reload

1. Run `air` from the project root

### Running with a static binary

1. From the project root, run `go build -o ./tmp/main .`
1. Run `./tmp/main`

## Future Improvements

- Refactor session.go. Need to break out some functionality into separate files and packages
- Deal with time zones
- Implement testing
- Additional reminder status options (e.g. "cancelled", "postponed", etc)
- Scheduler needs to account for new reminders. Currently only considers reminders that are stored in the DB
- Scheduler needs to take into account frequency instead of just session timestamp
- Make sure scheduler fires reminders at the correct time when the reminder is not in the past
- Deploy to Amazon ECS or DigitalOcean App Platform
- Document available endpoints
- Endpoint for querying the status of reminders
