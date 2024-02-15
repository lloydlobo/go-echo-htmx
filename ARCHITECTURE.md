# Architecture

- [Architecture](#architecture)
  - [The Problem: 4 Steps](#the-problem-4-steps)
    - [Purpose](#purpose)
    - [Function](#function)
    - [Usage](#usage)
    - [Implementation/Form](#implementationform)
  - [**Figure 1:** Application Architecture](#figure-1-application-architecture)

## The Problem: 4 Steps

### Purpose

- Synchronize the presence of users in a gathering for a live event.
- Similar to an electoral ballot vote count.

### Function

- Editable by a limited number of people (admins).

### Usage

- CRUD operations for user entries.
- Mark users as inactive if they leave, otherwise mark them as active.

### Implementation/Form

- Server holds an in-memory repository (data of users).
- Frontend is an interface to interact with the server in an event-driven fashion.

---

## **Figure 1:** Application Architecture

This architecture closely follows the "onion model" principle,
> ... where each layer doesn't know about the layer above it, and each layer is responsible for a specific thing.
>> [Application Architecture from templ.guide](https://templ.guide/project-structure/project-structure/#application-architecture)

```mermaid
graph TD
    %% Processes HTTP requests
    %% Does not contain application logic itself
    %% Uses services that carry out application logic
    %% Takes the responses from services and uses components to render HTML
    %% Creates HTTP responses
    subgraph "HTTP Handler"
        HTTPHandler
    end

    HTTPHandler -->|uses| ContactService
    HTTPHandler -->|renders| Components

    subgraph "View"
        Components -->| uses| htmx
        subgraph "Hypermedia (event driven)"
            _hyperscript -->|Client side interactivity| Components
            AlpineJS -->|Client side interactivity| Components
            htmx -->|Carries out event driven AJAX| HTTPHandler
        end
    end


    %% Carries out application logic such as orchestrating API calls, or making database calls
    %% Does not do anything related to HTML or HTTP
    %% Is not aware of the specifics of database calls
    subgraph "Services"
        ContactService
        %% Service2 -->|Carries out application logic| DatabaseCode
    end

    %% Handles database activity such as inserting and querying records
    %% Ensures that the database representation (records) doesn't leak to the service layer
    subgraph "Database access code"
        ContactService --> |use| DatabaseCode
        %% Service2 -->|Carries out application logic| DatabaseCode
    end

    %% subgraph "Models"
    %%     DatabaseCode -->|use| Contact
    %% end

    DatabaseCode -->|uses| SQLite

    subgraph "Some DB"
        SQLite[Database]
    end

    %% A more complex application may have a models package containing plain structs that represent common data structures in the application, such as User.

    %% As per https://go.dev/wiki/CodeReviewComments#interfaces the HTTP handler defines the interface that it's expecting, rather than the service defining its own interface.
```
