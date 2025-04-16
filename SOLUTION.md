# Weather Data API Solution

## Overview

This project implements a high-performance API for handling weather data, designed with scalability and maintainability in mind. The solution provides several key features:

- Functionality for retrieving and storing weather data
- WebSocket support for real-time data updates
- MongoDB-based persistence layer for efficient data storage
- Field projection to minimize payload sizes
- Comprehensive test suite with benchmarks

## Architecture

The application follows clean architecture principles with distinct separation of concerns:

### Core Components

- **Model Layer**: Domain entity, validation logic
- **Handler Layer**: HTTP request processing, WebSocket connections
- **Service Layer**: Ingestion logic, data transformation and parsing, queries
- **Storage Layer**: Database-specific implementations
- **Config**: Data config

### Technologies Used

- **Go 1.21+**: Modern, efficient backend language
- **MongoDB**: NoSQL database for flexible data storage
- **Gorilla Mux**: HTTP routing with pattern matching
- **Gorilla WebSocket**: Efficient WebSocket implementation
- **Zap Logger**: High-performance structured logging
- **Testify**: Testing toolkit for assertions and mocks

## Example API Endpoint Response

```bash
# Single date retrieval

http://localhost:8080/api/v1/weather/2023-01-01

[
    {
        "date": "2023-01-01T00:00:00Z",
        "temperature": 22.5265793208688,
        "humidity": 75.53907328514052
    }
]

## Single date retrieval with field projection

http://localhost:8080/api/v1/weather/2023-01-01?fields=temperature

[
    {
        "date": "2023-01-01T00:00:00Z",
        "temperature": 22.5265793208688
    }
]

```

```bash
# Date range retrieval

http://localhost:8080/api/v1/weather?from=2023-01-01&to=2023-01-31

[
    {
        "date": "2023-01-01T00:00:00Z",
        "temperature": 22.5265793208688,
        "humidity": 75.53907328514052
    },
    {
        "date": "2023-01-02T00:00:00Z",
        "temperature": 19.409728879858783,
        "humidity": 31.02362917554585
    },
    {
        "date": "2023-01-03T00:00:00Z",
        "temperature": 19.122795900530953,
        "humidity": 88.02329508463318
    },
    {
        "date": "2023-01-04T00:00:00Z",
        "temperature": 16.522612484526494,
        "humidity": 66.90348123810577
    },
    {
        "date": "2023-01-05T00:00:00Z",
        "temperature": 22.399257384336742,
        "humidity": 63.14634353949799
    },
    {
        "date": "2023-01-06T00:00:00Z",
        "temperature": 27.04349862673403,
        "humidity": 47.75699001533567
    },
    {
        "date": "2023-01-07T00:00:00Z",
        "temperature": 16.93350678263109,
        "humidity": 85.757500294187
    },
    {
        "date": "2023-01-08T00:00:00Z",
        "temperature": 23.1094952769305,
        "humidity": 45.95433764020746
    },
    {
        "date": "2023-01-09T00:00:00Z",
        "temperature": 12.934507354263928,
        "humidity": 79.6887967930017
    },
    {
        "date": "2023-01-10T00:00:00Z",
        "temperature": 13.996132171354782,
        "humidity": 89.10652076022791
    },
    {
        "date": "2023-01-11T00:00:00Z",
        "temperature": 11.17015886780472,
        "humidity": 77.00379873088835
    },
    {
        "date": "2023-01-12T00:00:00Z",
        "temperature": 34.26828606926582,
        "humidity": 61.13939522318591
    },
    {
        "date": "2023-01-13T00:00:00Z",
        "temperature": 10.096508787756527,
        "humidity": 33.96445583110032
    },
    {
        "date": "2023-01-14T00:00:00Z",
        "temperature": 14.464499201441408,
        "humidity": 58.34482735026438
    },
    {
        "date": "2023-01-15T00:00:00Z",
        "temperature": 25.32166882792481,
        "humidity": 56.29535681830221
    },
    {
        "date": "2023-01-16T00:00:00Z",
        "temperature": 12.034239971332632,
        "humidity": 42.16776247121972
    },
    {
        "date": "2023-01-17T00:00:00Z",
        "temperature": 32.04741257742081,
        "humidity": 55.41525820285827
    },
    {
        "date": "2023-01-18T00:00:00Z",
        "temperature": 27.990503946057203,
        "humidity": 51.46547304482032
    },
    {
        "date": "2023-01-19T00:00:00Z",
        "temperature": 34.159749285947335,
        "humidity": 39.82105566909891
    },
    {
        "date": "2023-01-20T00:00:00Z",
        "temperature": 22.69088868101912,
        "humidity": 56.482448599949166
    },
    {
        "date": "2023-01-21T00:00:00Z",
        "temperature": 17.51009207896218,
        "humidity": 45.76799737946934
    },
    {
        "date": "2023-01-22T00:00:00Z",
        "temperature": 23.737514319881786,
        "humidity": 61.323745240935864
    },
    {
        "date": "2023-01-23T00:00:00Z",
        "temperature": 33.27046793244933,
        "humidity": 32.10960358297165
    },
    {
        "date": "2023-01-24T00:00:00Z",
        "temperature": 23.019035931046513,
        "humidity": 84.37388518723705
    },
    {
        "date": "2023-01-25T00:00:00Z",
        "temperature": 16.68017579655797,
        "humidity": 78.98185833119136
    },
    {
        "date": "2023-01-26T00:00:00Z",
        "temperature": 31.93496972935299,
        "humidity": 63.154879950630274
    },
    {
        "date": "2023-01-27T00:00:00Z",
        "temperature": 19.29796871281153,
        "humidity": 81.10851496543458
    },
    {
        "date": "2023-01-28T00:00:00Z",
        "temperature": 10.034583749974978,
        "humidity": 87.74370442864803
    },
    {
        "date": "2023-01-29T00:00:00Z",
        "temperature": 16.1921255623079,
        "humidity": 36.63133764315963
    },
    {
        "date": "2023-01-30T00:00:00Z",
        "temperature": 17.95583772942656,
        "humidity": 67.84990850458823
    },
    {
        "date": "2023-01-31T00:00:00Z",
        "temperature": 31.469436705797555,
        "humidity": 89.87964005613952
    }
]

# Date range retrieval with field projection

http://localhost:8080/api/v1/weather?from=2023-01-01&to=2023-01-31&fields=temperature

[
    {
        "date": "2023-01-01T00:00:00Z",
        "temperature": 22.5265793208688
    },
    {
        "date": "2023-01-02T00:00:00Z",
        "temperature": 19.409728879858783
    },
    {
        "date": "2023-01-03T00:00:00Z",
        "temperature": 19.122795900530953
    },
    {
        "date": "2023-01-04T00:00:00Z",
        "temperature": 16.522612484526494
    },
    {
        "date": "2023-01-05T00:00:00Z",
        "temperature": 22.399257384336742
    },
    {
        "date": "2023-01-06T00:00:00Z",
        "temperature": 27.04349862673403
    },
    {
        "date": "2023-01-07T00:00:00Z",
        "temperature": 16.93350678263109
    },
    {
        "date": "2023-01-08T00:00:00Z",
        "temperature": 23.1094952769305
    },
    {
        "date": "2023-01-09T00:00:00Z",
        "temperature": 12.934507354263928
    },
    {
        "date": "2023-01-10T00:00:00Z",
        "temperature": 13.996132171354782
    },
    {
        "date": "2023-01-11T00:00:00Z",
        "temperature": 11.17015886780472
    },
    {
        "date": "2023-01-12T00:00:00Z",
        "temperature": 34.26828606926582
    },
    {
        "date": "2023-01-13T00:00:00Z",
        "temperature": 10.096508787756527
    },
    {
        "date": "2023-01-14T00:00:00Z",
        "temperature": 14.464499201441408
    },
    {
        "date": "2023-01-15T00:00:00Z",
        "temperature": 25.32166882792481
    },
    {
        "date": "2023-01-16T00:00:00Z",
        "temperature": 12.034239971332632
    },
    {
        "date": "2023-01-17T00:00:00Z",
        "temperature": 32.04741257742081
    },
    {
        "date": "2023-01-18T00:00:00Z",
        "temperature": 27.990503946057203
    },
    {
        "date": "2023-01-19T00:00:00Z",
        "temperature": 34.159749285947335
    },
    {
        "date": "2023-01-20T00:00:00Z",
        "temperature": 22.69088868101912
    },
    {
        "date": "2023-01-21T00:00:00Z",
        "temperature": 17.51009207896218
    },
    {
        "date": "2023-01-22T00:00:00Z",
        "temperature": 23.737514319881786
    },
    {
        "date": "2023-01-23T00:00:00Z",
        "temperature": 33.27046793244933
    },
    {
        "date": "2023-01-24T00:00:00Z",
        "temperature": 23.019035931046513
    },
    {
        "date": "2023-01-25T00:00:00Z",
        "temperature": 16.68017579655797
    },
    {
        "date": "2023-01-26T00:00:00Z",
        "temperature": 31.93496972935299
    },
    {
        "date": "2023-01-27T00:00:00Z",
        "temperature": 19.29796871281153
    },
    {
        "date": "2023-01-28T00:00:00Z",
        "temperature": 10.034583749974978
    },
    {
        "date": "2023-01-29T00:00:00Z",
        "temperature": 16.1921255623079
    },
    {
        "date": "2023-01-30T00:00:00Z",
        "temperature": 17.95583772942656
    },
    {
        "date": "2023-01-31T00:00:00Z",
        "temperature": 31.469436705797555
    }
]


```

## Performance

Benchmarks demonstrate excellent performance characteristics:

### HTTP Endpoints

```bash
BenchmarkHTTPHandler_GetWeatherByDate-24 78,666 req/sec 14.6μs latency 10KB/req 103 allocs/op

BenchmarkHTTPHandler_IngestWeatherData-24 40,958 req/sec 27.5μs latency 22KB/req 173 allocs/op
```

These HTTP endpoint benchmarks show:

- The GET endpoint can handle **78,666 requests per second** with just **14.6 microseconds** of processing time
- The POST endpoint processes **40,958 requests per second** with a **27.5 microsecond** latency
- Memory usage is efficient at **10KB per GET request** and **22KB per POST request**
- The higher allocation count for POST requests (173 vs 103) reflects the additional work of parsing and validating input data

### Data Processing

```bash
BenchmarkIngestService_IngestSingle-24 143,996 ops/sec 7.5μs/op 5.3KB/op 54 allocs/op

BenchmarkIngestService_IngestFile-24 943 ops/sec 1.06ms/op 694.8KB/op 7004 allocs/op
```

The data processing benchmarks reveal:

- Single record processing achieves **143,996 operations per second** with just **7.5 microseconds** per operation
- Batch file processing handles **943 files per second**, with each file containing approximately 365 records
- Batch processing shows excellent efficiency, processing each record in approximately **2.9 microseconds** (1.06ms ÷ 365 records)
- Memory usage scales efficiently with batch size, using just **1.9KB per record** in batch mode vs 5.3KB for single records

These benchmarks indicate the application can handle high volumes of traffic with minimal latency, making it suitable for production workloads. For context, database operations typically take 1-10ms, so the application processing time (7.5-27.5μs) represents less than 3% of a typical end-to-end request cycle.

## Setup Instructions

### Prerequisites

- Go 1.21+
- MongoDB 5.0+
- Make (optional, for using Makefile commands)

### Installation

MongoDB credentials provided separately through Git Gist file.

```bash
# Clone the repository (ssh)
git clone git@github.com:francescorizzello94/senior-fullstack-engineer-takehome.git

# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run benchmarks
make benchmark
```
