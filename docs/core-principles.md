

# **Architecting a Modern MUD: A Framework for LLM-Powered Worlds in Go**

## **Architectural Vision & Core Principles**

The creation of a modern, LLM-powered Multi-User Dungeon (MUD) framework in Go requires a paradigm shift away from traditional, monolithic MUD architectures. Early MUDs often tightly coupled game logic, networking protocols like Telnet, and content definition in flat files into a single, complex process.1 This report outlines a modern architectural approach that explicitly separates these concerns, treating the MUD not as a singular program but as a distributed system. The goal is to produce a comprehensive, extensible, and open-source platform for creating dynamic online worlds.

### **The Three Pillars of the Framework**

The architecture is founded upon three core, decoupled subsystems, each with a distinct responsibility. This separation is fundamental to achieving the project's goals of flexibility, maintainability, and future-proofing.

* **The Stateful Game Engine:** This is the authoritative core of the system. It is responsible for managing the entire game world's state, processing all game logic, and serving as the single source of truth for the simulation. It operates independently of how clients connect or how content is defined.  
* **The Hybrid Communication Layer:** This subsystem acts as a protocol-agnostic interface between the game engine and the external world. Its primary function is to receive requests from clients, translate them into abstract commands for the engine, and broadcast events from the engine back to the appropriate clients. This layer defines a clear and stable API boundary.  
* **The Layered Content DSL:** This is the declarative system for defining all game content, from the static structure of a room and its objects to the dynamic logic of a quest trigger or an NPC's behavior. This pillar is designed to empower world-builders, including those without programming expertise, to create rich and interactive content.

### **Core Design Principles**

A set of non-negotiable tenets guides all subsequent architectural decisions, ensuring the framework is robust, scalable, and adaptable.

* **Modularity and Decoupling:** Each of the three pillars must operate with minimal knowledge of the others. The game engine, for example, should not be aware of gRPC or WebSocket specifics; it should only process abstract commands and generate abstract events. This principle is essential for testability, independent development, and long-term maintenance of the framework.2  
* **Extensibility through Interfaces:** The framework must be designed for extension, not modification. Core systems will be defined by Go interfaces, allowing developers to provide their own implementations. This compile-time plugin architecture is favored over Go's native dynamic plugin system, which has significant cross-platform limitations and strict build environment requirements, making it unsuitable for a widely distributable open-source project.3  
* **Protocol-First Design:** The communication contract between the server and any client will be rigorously defined using Protocol Buffers. This contract is the cornerstone of the architecture, enabling the future development of a diverse ecosystem of clients—including text-based, 2D, 3D, and passive viewer clients—without necessitating any changes to the core server logic.5 The choice of gRPC is not merely for performance; it is a strategic decision to foster this ecosystem by establishing a formal, language-agnostic API.  
* **Authoritative Server Model:** To ensure fairness, prevent cheating, and maintain a consistent world state, all game logic will be executed and validated exclusively on the server. Clients will send intent, and the server will determine the outcome. This is a standard and essential practice in modern multiplayer game architectures, as demonstrated by mature frameworks like Nakama.7

## **The Stateful Server Engine**

The server engine is the heart of the MUD, containing the complete, living simulation of the game world. Its design prioritizes performance, concurrency, and a clean separation between the game's state and its logic.

### **State Management in a Concurrent World**

* **In-Memory Representation:** For maximum performance and low-latency interaction, the entire active game world state will be held in memory. This state will be represented by a hierarchy of Go structs for all core entities, such as World, Area, Room, Player, NPC, and Item.  
* **Concurrency Strategy:** Go's powerful concurrency primitives are central to managing a multi-user environment efficiently.9 The architecture will employ a hybrid concurrency model. Fine-grained locking using  
  sync.RWMutex will be applied to individual entities where contention is localized (e.g., a single player's inventory). For inter-system communication and processing player actions, Go channels will be used. This message-passing approach avoids widespread, coarse-grained locks, ensuring that the actions of one player do not block the entire game state.10  
* **Entity Composition:** While a full Entity Component System (ECS) might be overly complex for an initial implementation, its principles will inform the design. Game objects will be defined as compositions of smaller, reusable components (e.g., a PositionComponent, HealthComponent, InventoryComponent). This promotes flexibility and avoids the rigid inheritance hierarchies common in older object-oriented designs.  
* **State Persistence:** Although the active state is in-memory, it must be durable. The engine will periodically snapshot the world state to a persistent store. For simplicity, this could be structured files like JSON or TOML. For greater scalability and robustness, a database like CockroachDB, as used by the Nakama game server, would be an excellent choice.11 This ensures the world can be restored after a server crash or restart.

### **The Game Loop as the Heartbeat**

The game loop is the architectural component that decouples the unpredictable, real-time actions of players from the predictable, simulated time of the game world. A naive server might process commands as they arrive, creating race conditions and giving an unfair advantage to players with lower network latency. The game loop solves this by queuing all incoming actions and processing them in deterministic batches. This ensures fairness and simplifies state management by making the core simulation effectively single-threaded within each update phase, confining complex concurrency handling to the system's I/O edges.

* **Fixed-Timestep Implementation:** The engine will be driven by a fixed-timestep game loop, a standard and robust pattern in game development.12 This ensures that game state updates occur at a predictable, constant rate (e.g., 10 times per second), regardless of server load fluctuations. This determinism is crucial for consistent physics, AI behavior, and overall game feel. A  
  time.Ticker in Go is a natural fit for implementing this pattern.13  
* **Phases of a Tick:** Each iteration of the loop, or "tick," will be divided into distinct, sequential phases:  
  1. **Process Input Queue:** The loop first dequeues and processes all player commands that have been received from the communication layer since the previous tick.  
  2. **Update Game State:** This is the core simulation phase. The engine runs AI logic for NPCs, executes active scripts, updates the status of ongoing effects (like spells or poisons), and resolves combat calculations.  
  3. **Process Output Queue:** All events generated during the update phase (e.g., "You hit the goblin for 5 damage," "The goblin dies," "A door creaks open") are collected. These events are then passed to the communication layer's output queue to be broadcast to the relevant clients.  
* **Managing Time:** The loop will track elapsed real time to maintain its fixed cadence. If a tick takes longer to process than the allotted time step, the engine can prioritize the integrity of the simulation by skipping the output/rendering phase for one or more ticks to catch up, a technique that prevents the game world from falling behind the players' clocks.12

### **Designing for Extensibility: A Plugin-Based System**

A successful open-source framework must be extensible. Go's standard plugin package, which loads shared object (.so) files dynamically, is ill-suited for this purpose due to its lack of Windows support and its strict requirement for matching build environments between the main application and the plugin.3 This fragility makes it a poor choice for a framework intended for wide adoption.  
A more robust approach is an interface-driven, compile-time plugin architecture. Core game systems, such as CombatSystem, MagicSystem, and QuestSystem, will be defined as Go interfaces. The framework will provide default implementations for these systems. Game creators can then develop their own custom implementations in separate packages. These custom packages are then registered within their project's main.go file and compiled directly into the final server binary. This registration pattern, inspired by Go's database/sql driver model and the architecture of the Caddy web server, offers maximum flexibility and type safety without the pitfalls of dynamic linking.4

## **The Hybrid Communication Layer**

This layer serves as the server's API boundary, defining how clients interact with the game engine. The choice of protocols is critical for supporting the immediate need for a text-based client and the future goal of accommodating rich visual clients.

### **Protocol Strategy: Unifying gRPC and WebSockets**

A hybrid protocol model is adopted, leveraging the distinct strengths of gRPC and WebSockets to create an efficient and robust communication system.6

* **gRPC for Authoritative Actions:** All player actions that intend to mutate the game state will be handled via unary gRPC Remote Procedure Calls (RPCs). These include commands like Move(direction), Say(message), Attack(target), and Get(item). Using gRPC for these actions provides several advantages: a strictly defined API contract, strong typing, automatic validation, and the performance benefits of HTTP/2, such as stream multiplexing.5  
* **WebSockets for Real-Time Feedback:** All real-time feedback and world events generated by the server will be pushed to the client over a persistent WebSocket connection. This is the ideal channel for server-initiated broadcasts like chat messages, combat logs, descriptions of other players' actions, and atmospheric messages. This push-based model is far more efficient than requiring the client to constantly poll the server for updates.20  
* **HTTP(S) for Auxiliary Services:** For non-real-time, request-response interactions, such as viewing leaderboards, checking server status, or managing user accounts, standard RESTful HTTP endpoints can be exposed. This can be easily achieved using a gRPC-gateway, which translates RESTful JSON API calls into gRPC messages, making the server's auxiliary functions readily accessible to standard web frontends.

The following table provides a concise comparison of the chosen protocols against the traditional Telnet approach, justifying the hybrid architecture.

| Feature | gRPC | WebSockets | Traditional Telnet |
| :---- | :---- | :---- | :---- |
| **Primary Use Case** | Authoritative client actions (Move, Attack) | Real-time server broadcasts (Chat, Events) | Raw, unstructured text stream |
| **Underlying Protocol** | HTTP/2 | TCP (upgraded from HTTP/1.1) | TCP |
| **Data Format** | Protocol Buffers (Binary) | JSON/Binary (Flexible) | Plain Text / ANSI |
| **Streaming** | Bidirectional, Client, Server | Full-duplex Bidirectional | Full-duplex Bidirectional |
| **API Contract** | Strictly defined in .proto | Application-defined | Implicit / Ad-hoc |
| **Performance** | High throughput, low overhead (binary) | Very low latency for small messages | High overhead (text parsing) |
| **Ecosystem Support** | Excellent (multi-language code-gen) | Excellent (native in all browsers) | Niche (MUD clients) |
| **Future-Proofing** | **High:** Ideal for visual/API clients | **Medium:** Excellent for web, less structured | **Low:** Legacy, not suitable for non-text clients |

### **Defining the Wire Protocol with Protocol Buffers**

The canonical contract for all client-server communication will be defined in a single mud\_service.proto file. This file is the single source of truth for the API.

* **Service Definition:** It will define the MUDService and all its RPC methods, such as rpc Login(LoginRequest) returns (LoginResponse).  
* **Message Structures:** It will define all the message structs used in requests and responses, like LoginRequest and MoveRequest.  
* **Event Structure:** It will define a generic GameEvent message, which will be the payload for all data sent over the WebSocket channel. This message will use a oneof field to encapsulate various event types, such as ChatMessage, CombatLog, RoomDescription, or PlayerStateUpdate. This ensures that all real-time data is also strongly typed and structured.  
* **Code Generation:** The protocol buffer compiler (protoc), along with the Go plugins (protoc-gen-go and protoc-gen-go-grpc), will be used to automatically generate the Go server interfaces, client stubs, and all necessary data serialization and deserialization code. This eliminates a significant amount of manual boilerplate code and prevents a wide class of data handling errors.5

### **Implementation Patterns in Go**

* **gRPC Server:** The implementation will involve creating a Go struct that satisfies the MUDServiceServer interface generated by protoc. Each RPC handler method will act as a thin adapter. Its role is to validate the incoming request, construct a corresponding command object, and place that command onto the game engine's input channel. It may then wait for a direct response from the engine or return immediately for actions that are processed asynchronously.  
* **WebSocket Hub:** A central "hub" component will manage all active WebSocket connections. This hub will run in its own goroutine and will be responsible for registering new clients, handling disconnections, and broadcasting GameEvent messages. It will listen on the game engine's output channel and route each event to the appropriate clients (e.g., broadcasting a chat message to all players within a specific room). The gorilla/websocket library is the de-facto standard and a robust choice for this implementation.22  
* **Engine-Network Boundary:** The decoupling between the engine and the communication layer is achieved through two primary Go channels: an inputChan for commands flowing from the network to the engine, and an outputChan for events flowing from the engine back to the network. This creates a clean, buffered, and asynchronous boundary that is central to the architecture's modularity.

## **The World Definition DSL: An Authoring Deep Dive**

The Domain-Specific Language (DSL) is the core of the framework's authoring experience, enabling world-builders to define game content declaratively. It is designed as a layered system where different file types handle distinct aspects of content definition, promoting a clean separation of concerns.25

### **The Three-Layer Content Model**

The DSL is composed of three synergistic file types:

1. **TOML files** define the *structure and data* of game entities—what they are.  
2. **Markdown files** provide the rich, formatted *descriptions* of these entities—how they are presented to the player.  
3. **Embedded expressions** define the *dynamic logic and behavior* of entities—how they react to conditions and events.

### **Structural Definition with TOML**

TOML is selected for its high human readability and its direct, unambiguous mapping to data structures, making it an excellent format for configuration-style data.27 The framework will define a clear schema for game entities.

* **rooms.toml:**  
  Ini, TOML  
  \[room.liminal\_space\]  
  name \= "A Liminal Space"  
  description\_file \= "rooms/liminal\_space.md"  
  exits \= { north \= "room.grand\_hall" }  
  npc\_ids \= \["npc.ghost"\]

* **npcs.toml:**  
  Ini, TOML  
  \[npc.shopkeeper\]  
  name \= "Bartholomew the Merchant"  
  description\_file \= "npcs/shopkeeper.md"  
  inventory \= \["item.iron\_sword", "item.health\_potion"\]

* **items.toml:**  
  Ini, TOML  
  \[item.iron\_sword\]  
  name \= "a simple iron sword"  
  description\_file \= "items/iron\_sword.md"  
  type \= "weapon"  
  properties \= { damage \= 10, weight \= 5 }

For parsing, the pelletier/go-toml library (specifically v2) is recommended due to its high performance, strict mode for validating files against Go structs, and detailed error reporting, which is invaluable for content creators.28

### **Rich Descriptions with Markdown**

Using Markdown for descriptions allows for much richer text formatting than the plain text files of traditional MUDs. This formatting can be rendered as ANSI escape codes for terminal clients or as HTML for web clients.  
Crucially, the framework will not simply render Markdown to a final format. Instead, it will parse the Markdown source into an Abstract Syntax Tree (AST). This intermediate representation allows the game engine to programmatically inspect and manipulate the content before it is shown to the player. For example, the engine could traverse the AST to find all nouns and make them automatically interactive, or it could dynamically insert contextual information (like the current weather) into a specific paragraph. This transforms static text into a dynamic, interactive surface. The yuin/goldmark library is the ideal choice for this task, given its strong compliance with the CommonMark standard, its high performance, and its extensible, AST-based architecture.30 Other libraries like  
gomarkdown/markdown are also viable alternatives.31

### **Dynamic Logic with expr**

A static world is uninteresting. To introduce dynamic behavior—doors that only open with a key, NPCs that only appear at night, traps that trigger under certain conditions—the DSL needs a scripting or expression layer.  
The expr-lang/expr library is perfectly suited for this role. It is a safe, non-Turing complete expression language that is designed to be fast, secure, and easily embeddable within a Go application.33 Its simple syntax makes it accessible to non-programmers. These expressions can be embedded directly within the TOML definitions:

Ini, TOML

\[door.treasury\_door\]  
is\_locked \= true  
unlock\_condition \= "player.has\_item('treasury\_key')"

\[npc.ghost\]  
spawn\_condition \= "world.time.is\_night()"

The game engine will compile these expression strings when the world is loaded. At runtime, it will evaluate them against a context object containing the current state of the player, world, and other relevant entities. The framework can also expose custom Go functions to the expr environment, such as a roll\_dice() function, further extending the capabilities of the DSL.34

### **The DSL Interpreter**

The DSL Interpreter is the Go module that orchestrates the entire world-loading process. Its responsibilities include:

1. **Loading and Parsing:** It recursively scans the game content directory, parsing all TOML files into their corresponding Go structs and parsing all linked Markdown files into their AST representations.  
2. **Hydration and Linking:** After the initial parse, it performs a linking pass. It resolves string-based identifiers (e.g., converting "npc.ghost" in a room's npc\_ids field into a direct memory pointer to the loaded ghost NPC object). It also compiles all expr strings into executable programs.  
3. **Finalization:** The output of the interpreter is a single, fully hydrated, and interconnected World object in memory, which is then passed to the game engine to begin the simulation.

## **The LLM as a Dynamic Storyteller**

Integrating Large Language Models (LLMs) elevates the MUD from a static, scripted world to a dynamic and adaptive experience. The key to successful integration is using the highly structured DSL as a grounding mechanism. An unconstrained LLM will hallucinate and fail to maintain a consistent state.35 The architecture presented here uses the DSL and the in-memory game state as the absolute source of truth, providing the LLM with a rich, factual context for every task it performs. The LLM is a creative partner, not the master of the world state.

### **LLM Service Abstraction**

To avoid vendor lock-in and maintain flexibility, the framework will define a generic LLMService interface in Go. This interface will declare abstract methods like GenerateText(ctx, prompt) and ParseIntent(ctx, input). Concrete implementations will then be created as adapters for various LLM providers (e.g., OpenAI, Anthropic, Google Gemini, or locally-hosted models via Ollama). The tmc/langchaingo library is an excellent choice for this, as it already provides a unified interface over numerous LLM providers.36 For direct access to a specific provider like OpenAI, the  
sashabaranov/go-openai client is a popular and robust option.38 Go's performance and strong concurrency model make it exceptionally well-suited for building production-grade LLM infrastructure.40

### **A Hybrid Command Parser**

Relying on an LLM to parse every single player command would introduce unacceptable network latency for simple, common actions.35 Therefore, a two-stage hybrid parser is proposed:

1. **Stage 1: Fast Verb-Noun Parser:** A simple, deterministic parser (using string splitting or regular expressions) will handle a predefined list of core game commands (look, get \<item\>, n, inv, etc.). This provides the instant feedback players expect for the majority of their interactions, following the traditional text adventure parsing model.41  
2. **Stage 2: LLM Intent Parser (Fallback):** If the fast parser does not recognize the command, the raw player input is passed to the LLMService.ParseIntent method. A carefully crafted prompt will instruct the LLM to analyze the natural language sentence and return a structured JSON object representing the player's intent. For example, the input "ask the bartender about any rumors he's heard" would be translated into {"command": "ask", "target": "bartender", "topic": "rumors"}. The game engine then executes this structured, unambiguous command.42

### **Prompt Engineering for Dynamic Content**

The static content defined in the DSL serves as the foundational context for LLM prompts, ensuring that generated content is grounded in the established game world.

* **Dynamic NPC Dialogue:** When a player initiates a conversation, the engine constructs a detailed prompt for the LLM. This prompt will include the NPC's core personality and knowledge (from its TOML/Markdown files), role-playing instructions ("You are a grumpy dwarven blacksmith..."), the recent conversation history, and relevant facts about the player's current state. This allows for dialogue that is both dynamically generated and deeply consistent with the game's lore.44  
* **Generative Descriptions:** The engine can also use LLMs to embellish static descriptions. For example, it can provide the base Markdown description of a room to an LLM and ask it to add sensory details based on the current game state, such as the weather, time of day, or the aftermath of a recent battle.47

### **Managing LLM Latency and State**

* **Asynchronous Generation:** LLM API calls, especially for non-critical content like an NPC's internal monologue, can be executed in a separate goroutine. This prevents the main game loop from blocking while waiting for the API response.  
* **Caching:** To reduce API costs and improve response times, results from the LLM for common prompts can be cached.  
* **Streaming Responses:** For generating longer pieces of text, the framework should stream the response from the LLM API token-by-token. Libraries like go-openai and litellm support this.39 These tokens can be forwarded in real-time over the client's WebSocket connection, providing the user with immediate feedback instead of a long, jarring pause.

## **The Framework Toolchain**

To transform this collection of components into a usable, open-source framework, a dedicated toolchain for authoring, validating, and deploying game projects is essential.

### **Authoring and Validation CLI**

A command-line interface (CLI) is the primary tool for developers and world-builders to interact with the framework. The spf13/cobra library is the industry standard for building powerful CLIs in Go and is the recommended choice.49 The CLI will provide several key commands:

* mud-framework new \<project\_name\>: Scaffolds a new game project directory, populating it with a standard folder structure and example TOML, Markdown, and configuration files.  
* mud-framework validate: A crucial tool for content creators. This command runs the DSL Interpreter in a "dry-run" mode. It performs comprehensive validation, checking for TOML syntax errors, broken links between files, invalid expr syntax, and orphaned entities (e.g., an item that is defined but never placed in the world). This provides immediate, actionable feedback.  
* mud-framework run: Compiles and runs the game server using the content from the current project directory, intended for local development and testing.  
* mud-framework build: Compiles a distributable server binary and bundles the game content.

### **Publishing and Deployment**

* **Content Bundling:** For easy distribution, the entire game content directory (TOML, Markdown, etc.) can be embedded directly into the final Go binary using Go's native embed package. This creates a single, self-contained executable.  
* **Dockerization:** A Dockerfile will be provided to build the Go application and create a container image. This is the standard for modern server deployment, ensuring a consistent runtime environment and simplifying orchestration.11  
* **Configuration Management:** The server will be configured using a combination of a config.toml file and environment variables, adhering to the 12-factor app methodology. The spf13/viper library is the natural companion to cobra for handling this layered configuration.49

### **The Reference Go Client**

The framework must ship with a functional, text-based client to serve as a reference implementation and provide an out-of-the-box playable experience. This Go-based CLI application will:

* Establish a gRPC connection to the server for sending player commands.  
* Establish a WebSocket connection for receiving real-time game events.  
* Use a terminal library like c-bata/go-prompt or chzyer/readline to provide a polished user experience with command history and line editing.51  
* Run two primary goroutines concurrently: one for reading user input from the terminal and sending it via gRPC, and another for continuously reading events from the WebSocket and printing them to the console.

## **Synthesis and Strategic Roadmap**

This report has detailed a comprehensive architecture for a modern, LLM-powered MUD framework in Go. The design prioritizes modularity, extensibility, and a protocol-first approach to ensure a robust and future-proof platform.

### **Summary of Technology Recommendations**

The success of the framework relies on leveraging the strengths of the vibrant Go ecosystem. The following table summarizes the key recommended libraries and the justification for their selection.

| Category | Recommended Library | Justification |
| :---- | :---- | :---- |
| **CLI Framework** | spf13/cobra | Industry standard, powerful subcommands, ecosystem integration (Viper).49 |
| **gRPC** | google.golang.org/grpc | Official library, robust, performant, foundational to the protocol-first design.5 |
| **WebSockets** | gorilla/websocket | De-facto standard, feature-rich, widely used and tested.22 |
| **TOML Parsing** | pelletier/go-toml (v2) | High performance, strict mode for validation, excellent error reporting.28 |
| **Markdown Parsing** | yuin/goldmark | CommonMark compliant, highly extensible, performant, AST-based processing.30 |
| **Expression Language** | expr-lang/expr | Safe, fast, simple syntax, easy to embed and extend with custom functions.33 |
| **LLM Abstraction** | tmc/langchaingo | Provides a unified interface over multiple LLM providers, reducing vendor lock-in.36 |

### **Path to a Visual Client**

The protocol-first design is the key enabler for future visual clients. Because all communication is structured and defined in the mud\_service.proto file, building a graphical client becomes a matter of creating a new "view" for the existing data stream. A developer could use the protocol buffer compiler to generate client code in C\# for Unity, C++ for Unreal Engine, or TypeScript for a web-based client using a library like Three.js. Frameworks like Nakama demonstrate the power of this approach, supporting a wide range of game engines with a single server backend.54 The visual client would connect to the same gRPC and WebSocket endpoints, receive the same structured  
GameEvent messages, and be responsible for translating that data (e.g., a RoomDescription event) into a graphical representation.

### **Open-Source Community Strategy**

For the framework to thrive as an open-source project, fostering a healthy community is paramount.

* **Documentation is Paramount:** Success hinges on comprehensive documentation for the DSL schema, the server's wire protocol, and the interfaces for the plugin system. Clear documentation, tutorials, and runnable examples lower the barrier to entry for new users and contributors.49  
* **Idiomatic Go Best Practices:** The codebase must adhere to established Go idioms: the use of small, focused interfaces; clear and concise naming conventions; robust error handling; and a comprehensive test suite. This makes the code easier for other Go developers to read, understand, and contribute to.57  
* **Contribution Guidelines:** The project should establish clear contribution guidelines, including a CONTRIBUTING.md file, issue and pull request templates, and a code of conduct. This provides a structured and welcoming environment for community engagement, as seen in successful open-source projects like GoMud and Nakama.9

#### **Works cited**

1. Long-time Dev Looking to Build a Community-Driven MUD \- Anyone Interested? \- Reddit, accessed September 4, 2025, [https://www.reddit.com/r/MUD/comments/1jbxv5c/longtime\_dev\_looking\_to\_build\_a\_communitydriven/](https://www.reddit.com/r/MUD/comments/1jbxv5c/longtime_dev_looking_to_build_a_communitydriven/)  
2. Go best practices for project : r/golang \- Reddit, accessed September 4, 2025, [https://www.reddit.com/r/golang/comments/13uwq5m/go\_best\_practices\_for\_project/](https://www.reddit.com/r/golang/comments/13uwq5m/go_best_practices_for_project/)  
3. Building Extensible Go Applications with Plugins | by Thisara Weerakoon \- Medium, accessed September 4, 2025, [https://medium.com/@thisara.weerakoon2001/building-extensible-go-applications-with-plugins-19a4241f3e9a](https://medium.com/@thisara.weerakoon2001/building-extensible-go-applications-with-plugins-19a4241f3e9a)  
4. Clean Architecture and Plugin Systems in Go: A Practical Example : r/golang \- Reddit, accessed September 4, 2025, [https://www.reddit.com/r/golang/comments/1hwioc3/clean\_architecture\_and\_plugin\_systems\_in\_go\_a/](https://www.reddit.com/r/golang/comments/1hwioc3/clean_architecture_and_plugin_systems_in_go_a/)  
5. Basics tutorial | Go \- gRPC, accessed September 4, 2025, [https://grpc.io/docs/languages/go/basics/](https://grpc.io/docs/languages/go/basics/)  
6. The Duel of Data: gRPC vs WebSockets \- Apidog, accessed September 4, 2025, [https://apidog.com/blog/grpc-vs-websockets/](https://apidog.com/blog/grpc-vs-websockets/)  
7. Getting Started \- Heroic Labs Documentation, accessed September 4, 2025, [https://heroiclabs.com/docs/nakama/](https://heroiclabs.com/docs/nakama/)  
8. Authoritative Multiplayer \- Heroic Labs Documentation, accessed September 4, 2025, [https://heroiclabs.com/docs/nakama/concepts/multiplayer/authoritative/](https://heroiclabs.com/docs/nakama/concepts/multiplayer/authoritative/)  
9. GoMudEngine/GoMud: A Go based MUD (Multi-User ... \- GitHub, accessed September 4, 2025, [https://github.com/GoMudEngine/GoMud](https://github.com/GoMudEngine/GoMud)  
10. Golang 10 Best Practices \- Codefinity, accessed September 4, 2025, [https://codefinity.com/blog/Golang-10-Best-Practices](https://codefinity.com/blog/Golang-10-Best-Practices)  
11. heroiclabs/nakama: Distributed server for social and realtime games and apps. \- GitHub, accessed September 4, 2025, [https://github.com/heroiclabs/nakama](https://github.com/heroiclabs/nakama)  
12. Game Loop · Sequencing Patterns, accessed September 4, 2025, [https://gameprogrammingpatterns.com/game-loop.html](https://gameprogrammingpatterns.com/game-loop.html)  
13. Game Loop simulation in Golang \- Stack Overflow, accessed September 4, 2025, [https://stackoverflow.com/questions/40696458/game-loop-simulation-in-golang](https://stackoverflow.com/questions/40696458/game-loop-simulation-in-golang)  
14. Plugins with Go. How to use Go's standard library to… | by Rafael Oliveira | ProFUSION Engineering | Medium, accessed September 4, 2025, [https://medium.com/profusion-engineering/plugins-with-go-7ea1e7a280d3](https://medium.com/profusion-engineering/plugins-with-go-7ea1e7a280d3)  
15. Idiomatic approach to a Go plugin-based system \- Stack Overflow, accessed September 4, 2025, [https://stackoverflow.com/questions/35708608/idiomatic-approach-to-a-go-plugin-based-system](https://stackoverflow.com/questions/35708608/idiomatic-approach-to-a-go-plugin-based-system)  
16. WebSocket vs gRPC: Performance Comparison for Enterprises \- Lightyear.ai, accessed September 4, 2025, [https://lightyear.ai/tips/websocket-versus-grpc-performance](https://lightyear.ai/tips/websocket-versus-grpc-performance)  
17. gRPC vs. WebSocket: Key differences and which to use \- Ably, accessed September 4, 2025, [https://ably.com/topic/grpc-vs-websocket](https://ably.com/topic/grpc-vs-websocket)  
18. gRPC vs WebSocket | When Is It Better To Use? \- Wallarm, accessed September 4, 2025, [https://www.wallarm.com/what/grpc-vs-websocket-when-is-it-better-to-use](https://www.wallarm.com/what/grpc-vs-websocket-when-is-it-better-to-use)  
19. What is the difference between grpc and websocket? Which one is more suitable for bidirectional streaming connection? \- Stack Overflow, accessed September 4, 2025, [https://stackoverflow.com/questions/46904674/what-is-the-difference-between-grpc-and-websocket-which-one-is-more-suitable-fo](https://stackoverflow.com/questions/46904674/what-is-the-difference-between-grpc-and-websocket-which-one-is-more-suitable-fo)  
20. Game Server : r/golang \- Reddit, accessed September 4, 2025, [https://www.reddit.com/r/golang/comments/1hix5ql/game\_server/](https://www.reddit.com/r/golang/comments/1hix5ql/game_server/)  
21. How To: Build a gRPC Server In Go | by Pascal Allen | Medium, accessed September 4, 2025, [https://pascalallen.medium.com/how-to-build-a-grpc-server-in-go-943f337c4e05](https://pascalallen.medium.com/how-to-build-a-grpc-server-in-go-943f337c4e05)  
22. Pumping Messages with Websocket \- Jon Brown's Webpage, accessed September 4, 2025, [https://brojonat.com/posts/websockets/](https://brojonat.com/posts/websockets/)  
23. websocket package \- github.com/gorilla/websocket \- Go Packages, accessed September 4, 2025, [https://pkg.go.dev/github.com/gorilla/websocket](https://pkg.go.dev/github.com/gorilla/websocket)  
24. How do you host concurrent websocket connections using Golang Gorilla/mux?, accessed September 4, 2025, [https://stackoverflow.com/questions/55333469/how-do-you-host-concurrent-websocket-connections-using-golang-gorilla-mux](https://stackoverflow.com/questions/55333469/how-do-you-host-concurrent-websocket-connections-using-golang-gorilla-mux)  
25. Getting Started with Domain-Specific Languages (DSLs) | Better Stack Community, accessed September 4, 2025, [https://betterstack.com/community/guides/scaling-python/dsl-fundamentals/](https://betterstack.com/community/guides/scaling-python/dsl-fundamentals/)  
26. Lingo: A Go micro language framework for building Domain Specific Languages \- GitLab, accessed September 4, 2025, [https://about.gitlab.com/blog/a-go-micro-language-framework-for-building-dsls/](https://about.gitlab.com/blog/a-go-micro-language-framework-for-building-dsls/)  
27. TOML: Tom's Obvious Minimal Language, accessed September 4, 2025, [https://toml.io/en/](https://toml.io/en/)  
28. pelletier/go-toml: Go library for the TOML file format \- GitHub, accessed September 4, 2025, [https://github.com/pelletier/go-toml](https://github.com/pelletier/go-toml)  
29. toml package \- github.com/pelletier/go-toml \- Go Packages, accessed September 4, 2025, [https://pkg.go.dev/github.com/pelletier/go-toml](https://pkg.go.dev/github.com/pelletier/go-toml)  
30. yuin/goldmark: :trophy: A markdown parser written in Go. Easy to extend, standard(CommonMark) compliant, well structured. \- GitHub, accessed September 4, 2025, [https://github.com/yuin/goldmark](https://github.com/yuin/goldmark)  
31. gomarkdown/markdown: markdown parser and HTML renderer for Go \- GitHub, accessed September 4, 2025, [https://github.com/gomarkdown/markdown](https://github.com/gomarkdown/markdown)  
32. Advanced markdown processing in Go, accessed September 4, 2025, [https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html](https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html)  
33. Expr | Expression language, accessed September 4, 2025, [https://expr-lang.org/](https://expr-lang.org/)  
34. rosbit/go-expr: make expr-lang be embedded easily in Golang \- GitHub, accessed September 4, 2025, [https://github.com/rosbit/go-expr](https://github.com/rosbit/go-expr)  
35. Getting an LLM to Play Text Adventures \- Entropic Thoughts, accessed September 4, 2025, [https://entropicthoughts.com/getting-an-llm-to-play-text-adventures](https://entropicthoughts.com/getting-an-llm-to-play-text-adventures)  
36. llms package \- github.com/tmc/langchaingo/llms \- Go Packages, accessed September 4, 2025, [https://pkg.go.dev/github.com/tmc/langchaingo/llms](https://pkg.go.dev/github.com/tmc/langchaingo/llms)  
37. 10 Go Libraries Every LLM Learner Needs to Master Now. | by Nayak Satya | Medium, accessed September 4, 2025, [https://medium.com/@nayaksatya2012/10-go-libraries-every-llm-learner-needs-to-master-now-a7425fc2b1b7](https://medium.com/@nayaksatya2012/10-go-libraries-every-llm-learner-needs-to-master-now-a7425fc2b1b7)  
38. sashabaranov/go-openai: OpenAI ChatGPT, GPT-5, GPT-Image-1, Whisper API clients for Go \- GitHub, accessed September 4, 2025, [https://github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai)  
39. openai package \- github.com/sashabaranov/go-openai \- Go Packages, accessed September 4, 2025, [https://pkg.go.dev/github.com/sashabaranov/go-openai](https://pkg.go.dev/github.com/sashabaranov/go-openai)  
40. Scaling LLMs with Golang: How we serve millions of LLM requests \- Assembled, accessed September 4, 2025, [https://www.assembled.com/blog/scaling-llms-with-golang-how-we-serve-millions-of-llm-requests](https://www.assembled.com/blog/scaling-llms-with-golang-how-we-serve-millions-of-llm-requests)  
41. How should I parse user input in a text adventure game?, accessed September 4, 2025, [https://gamedev.stackexchange.com/questions/27004/how-should-i-parse-user-input-in-a-text-adventure-game](https://gamedev.stackexchange.com/questions/27004/how-should-i-parse-user-input-in-a-text-adventure-game)  
42. Using generative AI for sub-tasks in text adventures? \- Development Systems, accessed September 4, 2025, [https://intfiction.org/t/using-generative-ai-for-sub-tasks-in-text-adventures/75986](https://intfiction.org/t/using-generative-ai-for-sub-tasks-in-text-adventures/75986)  
43. Intra: design notes on an LLM-driven text adventure \- Ian Bicking, accessed September 4, 2025, [https://ianbicking.org/blog/2025/07/intra-llm-text-adventure](https://ianbicking.org/blog/2025/07/intra-llm-text-adventure)  
44. Unleashing the potential of prompt engineering for large language models \- PMC, accessed September 4, 2025, [https://pmc.ncbi.nlm.nih.gov/articles/PMC12191768/](https://pmc.ncbi.nlm.nih.gov/articles/PMC12191768/)  
45. Guidance for Dynamic Non-Player Character (NPC) Dialogue on AWS, accessed September 4, 2025, [https://aws.amazon.com/solutions/guidance/dynamic-non-player-character-dialogue-on-aws/](https://aws.amazon.com/solutions/guidance/dynamic-non-player-character-dialogue-on-aws/)  
46. Using LLMs for Game Dialogue: Smarter NPC Interactions \- Pyxidis, accessed September 4, 2025, [https://pyxidis.tech/llm-for-dialogue](https://pyxidis.tech/llm-for-dialogue)  
47. YOU SEE AN LLM HERE: Integrating Language Models Into Your Text Adventure Games, accessed September 4, 2025, [https://machinelearningmastery.com/you-see-an-llm-here-integrating-language-models-text-adventure-games/](https://machinelearningmastery.com/you-see-an-llm-here-integrating-language-models-text-adventure-games/)  
48. BerriAI/litellm: Python SDK, Proxy Server (LLM Gateway) to call 100+ LLM APIs in OpenAI format \- \[Bedrock, Azure, OpenAI, VertexAI, Cohere, Anthropic, Sagemaker, HuggingFace, Replicate, Groq\] \- GitHub, accessed September 4, 2025, [https://github.com/BerriAI/litellm](https://github.com/BerriAI/litellm)  
49. The CLI Framework Developers Love | Cobra: A Commander for Modern CLI Apps, accessed September 4, 2025, [https://cobra.dev/](https://cobra.dev/)  
50. spf13/cobra: A Commander for modern Go CLI interactions \- GitHub, accessed September 4, 2025, [https://github.com/spf13/cobra](https://github.com/spf13/cobra)  
51. Command-line Interfaces (CLIs) \- The Go Programming Language, accessed September 4, 2025, [https://go.dev/solutions/clis](https://go.dev/solutions/clis)  
52. Pterodactyl, accessed September 4, 2025, [https://pterodactyl.io/](https://pterodactyl.io/)  
53. Evaluating GoLang CLI Packages. Comparing survey, promptui, and… \- Thomas Jay Rush, accessed September 4, 2025, [https://tjayrush.medium.com/evaluating-golang-cli-packages-2ae34bb79787](https://tjayrush.medium.com/evaluating-golang-cli-packages-2ae34bb79787)  
54. Nakama C/C++ Client SDK \- GitHub Pages, accessed September 4, 2025, [https://heroiclabs.github.io/nakama-cpp/](https://heroiclabs.github.io/nakama-cpp/)  
55. @heroiclabs/nakama-js-base \- GitHub Pages, accessed September 4, 2025, [https://heroiclabs.github.io/nakama-js/](https://heroiclabs.github.io/nakama-js/)  
56. nakama \- Dart API docs \- Pub.dev, accessed September 4, 2025, [https://pub.dev/documentation/nakama/latest/](https://pub.dev/documentation/nakama/latest/)  
57. styleguide | Style guides for Google-originated open-source projects, accessed September 4, 2025, [https://google.github.io/styleguide/go/best-practices.html](https://google.github.io/styleguide/go/best-practices.html)  
58. Best Practices for a New Go Developer \- CloudBees, accessed September 4, 2025, [https://www.cloudbees.com/blog/best-practices-for-a-new-go-developer](https://www.cloudbees.com/blog/best-practices-for-a-new-go-developer)  
59. Go Coding Official Standards and Best Practices \- DEV Community, accessed September 4, 2025, [https://dev.to/leapcell/go-coding-official-standards-and-best-practices-284k](https://dev.to/leapcell/go-coding-official-standards-and-best-practices-284k)  
60. Nakama: The leading open source game server for studios and publishers \- Heroic Labs, accessed September 4, 2025, [https://heroiclabs.com/nakama/](https://heroiclabs.com/nakama/)