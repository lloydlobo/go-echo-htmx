# Architecture

## **Figure 1:** Application Architecture

This architecture follows a typical "onion model" closely.
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

    HTTPHandler -->|uses| Service1
    HTTPHandler -->|renders| Components

    subgraph "View"
        Components -->| uses| htmx
        subgraph "Hypermedia (event driven)"
            _hyperscript -->|Client side interactivity| Components
            htmx -->|Carries out event driven AJAX| HTTPHandler
        end
    end


    %% Carries out application logic such as orchestrating API calls, or making database calls
    %% Does not do anything related to HTML or HTTP
    %% Is not aware of the specifics of database calls
    subgraph "Services"
        Service1
        %% Service2 -->|Carries out application logic| DatabaseCode
    end

    %% Handles database activity such as inserting and querying records
    %% Ensures that the database representation (records) doesn't leak to the service layer
    subgraph "Database access code"
        Service1 --> |use| DatabaseCode
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
