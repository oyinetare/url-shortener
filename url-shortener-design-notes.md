# URL Shortener System Design

## Requirements

### Functional Requirements

#### System Overview

- **Inputs/Outputs**: Pass in long URL, return short URL
- **Core Functionality**: Short URL redirects to long URL
- **Traffic Volume**: Define expected number of users
- **Usage Patterns**:
  - Analytics are important
  - Read-heavy system (fewer writes)
  - URLs cannot be deleted or updated
  - URLs expire after one month
  - Short URLs use character set: `[0-9, a-z, A-Z]`
  - Short URLs are 8 characters long
  - Short URLs should fit into 64-bit storage
  - Deduplication required

#### Key Questions

- How much data do we expect to handle?
- How many requests per second do we expect?
- What is the expected read-to-write ratio?

#### Out of Scope

- Not storing user information
- Authentication

### Non-Functional Requirements

- Scalable
- High latency tolerance
- Low throughput requirements
- Maintainable
- Testable
- High availability
- Fault tolerant

## Back-of-the-Envelope Calculations

### Metrics to Calculate

- **Throughput**: QPS for read and write queries
- **Latency**: Expected latency for read and write queries
- **Read/Write Ratio**: Determine the ratio
- **Traffic Estimates**:
  - Write (QPS, Volume of data)
  - Read (QPS, Volume of data)
- **Storage Estimates**: Total storage needed
- **Memory Estimates**:
  - Cache data requirements
  - RAM and machine requirements
  - Disk/SSD storage requirements

### Capacity

- Total possibilities: 62^8 (using alphanumeric characters)

## High-Level Design

### Components

#### Client Layer

- **API Communication**:
  - REST API (chosen over RPC for external use, GraphQL not needed)
  - `POST /api/v1/shorten`
    - Parameter: `longUrl`
  - `GET /api/v1/shortUrl`
    - Redirects to `longUrl`

#### Services

##### Shorten Service

**Logic**:

```
if URL in database/storage:
    return existing short URL
else:
    generate short_code
    store in database/storage
    return new short URL
```

**Short Code Generation Options**:

- Multi-master replication (auto-increment)
- UUID
- Ticket server
- Twitter Snowflake

**Hashing Options**:

- Hash with collision handling (MD5, SHA-1, CRC32)

##### Redirect Service

**Logic**:

```
if short URL in database:
    redirect to long URL
    return 302 status code (for analytics)
else:
    return error
```

#### Storage

- Initial design: Hash table on disk
- Schema:
  ```
  | short_url | long_url |
  |-----------|----------|
  ```

#### Testing

- Unit tests

### Problems with Initial Design

1. **Database bottleneck**: Constantly hitting DB/storage for URLs
   - **Solution**: Implement caching
2. **Scalability**: Hash map won't scale well
   - **Solution**: Use MySQL table with Bloom filter
3. **ID Generation**: Snowflake is best option
4. **Collision**: Possibility with hashing functions
5. **Performance**: 302 redirects impact server load

## Deep Dive & Scaling

### Improved Design

#### Components

- **Services**:
  - Shorten service
  - Redirect service
- **Cache Layer**: For frequently accessed URLs
- **Database**:

  ```
  | short_url | long_url | clicks_count |
  |-----------|----------|--------------|

  Primary Key: short_url
  ```

- **Testing**:
  - Integration tests
  - Unit tests

### Problems with Improved Design

- Locking and synchronization issues

### Final Scalable Design

#### Architecture Components

- **Client & Services Layer**:
  - API Gateway
  - Rate Limiter
  - REST APIs
  - CDN
  - DNS
  - Caching layer
  - Load Balancer / Reverse Proxy
  - Asynchronous processing (Message Queues)
- **Cache Layer**: Redis/Memcached

- **Database Layer**:
  - SQL database with:
    - Sharding
    - Indexing
    - Partitioning

### Additional Considerations

- Dockerize as microservices?
