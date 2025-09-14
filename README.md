# Finance App API

A REST API for managing financial transactions built with Go and Chi router (migrated from Gin).

## Features

- ✅ **Chi Router**: Lightweight and idiomatic HTTP router for Go
- ✅ **MongoDB Integration**: Persistent data storage with MongoDB
- ✅ **CRUD Operations**: Complete Create, Read, Update, Delete operations for transactions
- ✅ **Middleware**: Request logging, recovery, timeout, CORS, and request ID middleware
- ✅ **Error Handling**: Comprehensive error handling with consistent API responses
- ✅ **Validation**: Input validation for all API endpoints
- ✅ **Structured Logging**: JSON formatted logs using Logrus

## Bug Fixes & Improvements

### Fixed Bugs:
1. **Server error handling**: Fixed unreachable error handling code after `http.ListenAndServe`
2. **MongoDB disconnection**: Added proper context timeout for MongoDB disconnection
3. **Environment variable loading**: Added graceful handling when .env file is missing

### Improvements:
1. **Migrated from Gin to Chi**: Replaced Gin framework with Chi router for better performance and smaller footprint
2. **Enhanced middleware stack**: Added request ID, logging, recovery, timeout, and CORS middleware
3. **Consistent API responses**: Implemented standardized JSON response format across all endpoints
4. **Better error messages**: Improved error handling with descriptive error messages
5. **Input validation**: Added comprehensive validation for transaction creation and updates
6. **Repository improvements**: Enhanced database operations with better error handling

## Prerequisites

- Go 1.20 or higher
- MongoDB (running locally or remotely)
- Environment variables configured

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd axis_task_in_go
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the root directory:
```env
MONGO_URI=mongodb://localhost:27017
```

4. Build the application:
```bash
go build -o finance_app.exe ./src/cmd
```

5. Run the application:
```bash
./finance_app.exe
```

The server will start on port 1234.

## API Endpoints

### Health Check
- **GET** `/api/v1/health`
  - Returns the health status of the API
  - Response:
    ```json
    {
      "message": "OK",
      "timestamp": "2024-01-14T10:30:00Z",
      "status": "healthy"
    }
    ```

### Transactions

#### Get All Transactions
- **GET** `/api/v1/transactions`
  - Retrieves all transactions (sorted by newest first)
  - Response:
    ```json
    {
      "success": true,
      "message": "Transactions fetched successfully",
      "data": [...]
    }
    ```

#### Create Transaction
- **POST** `/api/v1/transactions`
  - Creates a new transaction
  - Request Body:
    ```json
    {
      "transactionType": "DEPOSIT",
      "amount": 100.50,
      "balance": 1000.50,
      "accountId": "507f1f77bcf86cd799439011"
    }
    ```
  - Response:
    ```json
    {
      "success": true,
      "message": "Transaction created successfully",
      "data": {...}
    }
    ```

#### Get Transaction by ID
- **GET** `/api/v1/transactions/{id}`
  - Retrieves a specific transaction by ID
  - Response:
    ```json
    {
      "success": true,
      "message": "Transaction fetched successfully",
      "data": {...}
    }
    ```

#### Get Transactions by Account ID
- **GET** `/api/v1/transactions/account/{accountId}`
  - Retrieves all transactions for a specific account
  - Response:
    ```json
    {
      "success": true,
      "message": "Transactions fetched successfully",
      "data": [...]
    }
    ```

## Transaction Types

- `DEPOSIT`: Money deposited into an account
- `WITHDRAW`: Money withdrawn from an account
- `TRANSFER`: Money transferred between accounts

## Response Format

All API responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...}
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error description"
}
```

## Project Structure

```
axis_task_in_go/
├── src/
│   ├── cmd/
│   │   └── server.go           # Main application entry point
│   ├── handlers/
│   │   └── app.go              # Application handler with dependencies
│   ├── models/
│   │   └── transactions.go     # Transaction data model
│   ├── repositories/
│   │   └── transactions.go     # Database operations
│   ├── routes/
│   │   └── index.go            # Route definitions and middleware
│   ├── services/
│   │   └── transactions.go     # Business logic and HTTP handlers
│   └── utils/
│       └── db.go               # Database connection utilities
├── .env                        # Environment variables
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
└── README.md                   # This file
```

## Middleware Stack

The application uses the following middleware (in order):
1. **Request ID**: Generates unique ID for each request
2. **Logger**: Logs all HTTP requests
3. **Recoverer**: Recovers from panics gracefully
4. **Timeout**: Sets 60-second timeout for requests
5. **CORS**: Enables Cross-Origin Resource Sharing

## Error Handling

The application provides detailed error messages for common scenarios:
- Invalid request body format
- Missing required fields
- Invalid ObjectID format
- Transaction not found
- Database connection errors
- Validation errors (amount must be > 0, valid transaction types, etc.)

## Development

### Running Tests
```bash
go test ./...
```

### Running with Hot Reload
Install Air for hot reloading:
```bash
go install github.com/air-verse/air@latest
air
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.
