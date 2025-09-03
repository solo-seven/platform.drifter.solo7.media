# Drifter Platform RPG

A multiplayer RPG platform built with Go, implementing a clean domain-driven architecture with client-server communication over WebSocket protocol.

## Architecture Overview

This project implements the RPG domain model defined in the technical specifications, featuring:

- **Domain Layer**: Pure business logic with entity-component system
- **Network Layer**: WebSocket-based client-server communication with protobuf serialization
- **Server Layer**: Game server with entity management, rules engine, and world state management
- **Client Layer**: CLI client for testing and interaction

## Project Structure

```
├── cmd/
│   ├── client/          # CLI client application
│   └── server/          # Game server application
├── internal/
│   ├── domain/          # Core domain interfaces and types
│   ├── network/         # Network protocol implementation
│   └── server/          # Server implementation
├── tests/
│   ├── unit/            # Unit tests
│   ├── integration/     # Integration tests
│   └── uat/             # User acceptance tests
├── proto/               # Protocol buffer definitions
├── generated/           # Generated protobuf code
└── config.yaml         # Server configuration
```

## Features

### Core Systems
- **Entity-Component System**: Flexible entity management with component-based architecture
- **Rules Engine**: Event-driven game rules with condition evaluation
- **World State Management**: Authoritative server-side state management
- **Network Protocol**: Efficient WebSocket communication with protobuf serialization

### Client Features
- **Interactive CLI**: Command-line interface for player interaction
- **Real-time Communication**: WebSocket-based messaging
- **Command System**: Movement, combat, chat, and admin commands

### Server Features
- **Connection Management**: WebSocket connection handling with heartbeat monitoring
- **Message Processing**: Player input, chat, and system message handling
- **State Synchronization**: Real-time world state updates
- **Graceful Shutdown**: Clean server shutdown with connection cleanup

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make (for build automation)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd platform.drifter.solo7.media
```

2. Install dependencies:
```bash
make deps
```

3. Install development tools:
```bash
make install-tools
```

### Building

Build all components:
```bash
make build
```

Build individual components:
```bash
make build-server
make build-client
```

### Running

Start the server:
```bash
make run-server
# or
./build/drifter-server
```

Start the client:
```bash
make run-client
# or
./build/drifter-client
```

### Development

Run with hot reload (requires air):
```bash
make dev-server
```

## Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-integration
```

### User Acceptance Tests
```bash
make test-uat
```

### Test Coverage
```bash
make test-coverage
```

### BDD Tests (Ginkgo)
```bash
make test-bdd
```

## Configuration

The server can be configured via `config.yaml`:

```yaml
server:
  port: 8080
  max_connections: 1000
  heartbeat_interval: 30s
  region_size: 1000.0
  max_entities_per_region: 1000
  log_level: info
```

## Client Commands

The CLI client supports the following commands:

- `/help` - Show help message
- `/move <direction>` - Move character (forward, back, left, right)
- `/attack <target>` - Attack a target
- `/chat <message>` - Send chat message
- `/quit` - Disconnect from server

You can also type regular messages to send them as chat messages.

## Protocol

The client-server communication uses WebSocket with protobuf message serialization. Message types include:

- **Player Input**: Movement, combat, and interaction commands
- **Chat Messages**: Player-to-player communication
- **State Updates**: Entity and world state changes
- **System Notifications**: Server announcements and status updates
- **Heartbeat**: Connection health monitoring

## Development Workflow

This project follows Test-Driven Development (TDD) methodology:

1. **Write Tests First**: Define behavior through tests
2. **Implement Features**: Write minimal code to pass tests
3. **Refactor**: Improve code while maintaining test coverage
4. **Integration Testing**: Test component interactions
5. **User Acceptance Testing**: Validate end-to-end scenarios

### Test Structure

- **Unit Tests**: Test individual components and functions
- **Integration Tests**: Test client-server communication
- **UAT Tests**: Test complete user scenarios

## Contributing

1. Write tests for new features
2. Implement features to pass tests
3. Ensure all tests pass: `make test`
4. Run integration tests: `make test-integration`
5. Run UAT tests: `make test-uat`

## License

[Add your license information here]

## Roadmap

### Phase 1: Core Framework ✅
- [x] Domain interfaces and data structures
- [x] WebSocket protocol implementation
- [x] Basic server and client
- [x] Test framework setup

### Phase 2: Game Master Definition
- [ ] Core mechanics repository
- [ ] Action resolution system
- [ ] Rule precedence framework
- [ ] Conflict resolution

### Phase 3: Player's Guide Foundation
- [ ] Character creation system
- [ ] Skill definitions
- [ ] Advancement model
- [ ] Combat mechanics

### Phase 4: Content Expansion
- [ ] Monster Manual structure
- [ ] World book templates
- [ ] Environmental systems
- [ ] Cultural context framework

### Phase 5: Integration and Polish
- [ ] Complete cross-reference validation
- [ ] Advanced authoring tools
- [ ] Module system implementation
- [ ] Publishing pipeline

