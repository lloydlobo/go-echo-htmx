# Architecture

## **Figure 1:** Application Architecture

This architecture follows a typical "onion model" closely.
> ... where each layer doesn't know about the layer above it, and each layer is responsible for a specific thing.
>> [Application Architecture from templ.guide](https://templ.guide/project-structure/project-structure/#application-architecture)

```mermaid
graph TD
    subgraph "HTTP Handler"
        HTTPHandler
        %% HTTPHandler -->|Processes HTTP requests| Response
        %% Response -->|Creates HTTP responses| Components
    end

    subgraph "View"
        Components -->| uses| htmx
        subgraph "Hypermedia (event driven)"
            _hyperscript -->|Client side interactivity| Components
            htmx -->|Carries out event driven AJAX| HTTPHandler
        end
    end

    subgraph "Services"
        Service1 -->|Carries out application logic| DatabaseCode
        %% Service2 -->|Carries out application logic| DatabaseCode
    end

    subgraph "Database access code"
        %%DatabaseCode -->|Handles database activity| SQLite
        DatabaseCode -->|uses| SQLite
    end

    subgraph "SQLite"
        SQLite[Database]
    end

    HTTPHandler -->|uses| Service1
    HTTPHandler -->|renders| Components

    %% HTTPHandler -->|uses| Service2
    %% HTTPHandler -->|renders| Components
    %% Service1 -->|use| DatabaseCode
    %% Service2 -->|use| DatabaseCode
    %%DatabaseCode -->|uses| SQLite;
```
