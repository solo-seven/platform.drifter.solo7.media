# **LLM-Powered MUD: Architecture & Domain Model**

This document outlines the technical components, domain model, and an initial user story map for creating a flexible, open-source framework for text-based role-playing games, powered by a Large Language Model (LLM).

## **Part 1: Research & Technical Components**

Based on your requirements, the architecture will be built in Go, prioritizing a clean separation between the game logic and the communication layer to support future visual clients.

### **1\. Server & Client Communication: gRPC**

For the wire protocol, **gRPC** is an excellent choice. It's a high-performance, open-source RPC framework that works across many languages.

* **Why gRPC?**  
  * **Contract-First:** You define your services and messages in .proto (Protocol Buffer) files. This creates a clear, strongly-typed contract between the server and any client (text, visual, etc.).  
  * **Performance:** It uses HTTP/2 for transport, making it much more efficient than traditional REST+JSON for streaming data and low-latency communication, which is ideal for a game.  
  * **Streaming:** gRPC has built-in support for bi-directional streaming. The server can stream game state updates, messages, and descriptions to the client, and the client can stream player commands.  
* **Recommended Go Library:** google.golang.org/grpc is the official Go implementation.

### **2\. The Domain Specific Language (DSL)**

The heart of the authoring system is a layered DSL that separates structure, logic, and descriptive content. This makes world-building intuitive and powerful.

#### **a. TOML for Structure**

TOML is perfect for defining the static elements of your game world. Its clear, hierarchical syntax is human-readable and less prone to errors than JSON or YAML for this use case.

* **Usage:** Define rooms, items, NPCs, and their base properties.  
* **Example room.toml:**  
  id \= "starter\_forest"  
  name \= "Whispering Glade"

  \# Link to the markdown file for the long description  
  description\_file \= "rooms/whispering\_glade.md" 

  \[\[exits\]\]  
  direction \= "north"  
  target\_room \= "deep\_woods"  
  description \= "A dark path leads north into the deeper woods."

  \[\[exits\]\]  
  direction \= "south"  
  target\_room \= "village\_gate"  
  description \= "A well-trodden path leads south toward the village."

  \[\[items\]\]  
  item\_id \= "common\_sword" \# Refers to a definition in an item file  
  spawn\_chance \= 0.1

  \[\[npcs\]\]  
  npc\_id \= "lost\_merchant"

* **Recommended Go Library:** github.com/pelletier/go-toml is a robust and widely-used library for parsing TOML files.

#### **b. Expression Language for Logic**

To handle dynamic game rules, triggers, and conditions, you need an embedded expression language. This avoids having to recompile the server for every logic change.

* **Usage:** Checking conditions (player has key, quest is active), scripting NPC behavior, validating commands.  
* **Example Logic in TOML:**  
  \# In an item definition for a locked chest  
  \[on\_use\]  
  \# The 'expr' engine will evaluate this script  
  script \= """  
      player.inventory.has("key\_of\_ancients") && \!self.flags.get("is\_open")  
  """  
  success\_message \= "The ancient key turns in the lock with a satisfying click."  
  failure\_message \= "The chest is sealed shut."

* **Recommended Go Library:** github.com/expr-lang/expr is a fantastic choice. It's specifically designed for Go, is memory-safe, strongly-typed, and terminates, preventing infinite loops—all critical features for a server running user-defined logic.

#### **c. Markdown for Descriptions**

For rich, multi-paragraph descriptions with formatting, Markdown is the standard.

* **Usage:** Room descriptions, item examinations, quest text. The game server will parse the Markdown and convert it to the appropriate format for the client (e.g., ANSI escape codes for a terminal, HTML for a web client).  
* **Recommended Go Library:** github.com/gomarkdown/markdown is a fast, compliant, and extensible Markdown parser for Go.

### **3\. LLM Integration**

The LLM should be treated as an external service that the core game engine calls. This keeps the game deterministic while allowing for dynamic, AI-generated content.

* **Architecture:** Create a dedicated llm service/package in your Go application. This service will handle:  
  * Constructing prompts based on game state (e.g., player context, NPC personality, recent events).  
  * Making asynchronous API calls to the LLM.  
  * Parsing the LLM's response.  
  * Managing API keys and rate limiting.  
* **Use Cases:**  
  * **Dynamic NPC Dialogue:** Instead of pre-scripted lines, the LLM can generate dialogue based on the NPC's personality and the player's questions.  
  * **Evolving Room Descriptions:** "You enter the forest. The wind howls, and you notice fresh wolf tracks in the mud, leading north."  
  * **Player Assistance:** A "hint" command could be powered by the LLM.

## **Part 2: Domain-Driven Design (DDD) Model**

DDD helps manage the complexity of a project like this by focusing on the core "domain"—the game itself. We'll define a **Ubiquitous Language** (a shared vocabulary) and model the system around it.

### **Core Concepts of Our Ubiquitous Language**

* **World:** The entire game universe.  
* **Area:** An outdoor location with continuous boundaries to other Areas (e.g., a forest, a plain, a desert).  
* **Structure:** An enclosed space with discrete entrances and exits (e.g., a building, a cave, a room).  
* **Location:** A generic term for a place a character can be, which can be either an Area or a Structure.  
* **Character:** The player's avatar in the game.  
* **NPC:** A Non-Player Character controlled by the system or LLM.  
* **Item:** An object that can be held, used, or found in the world.  
* **Action:** A command a Character can perform (e.g., move, get, talk).  
* **State:** The current condition of any entity (e.g., Character's health, an Area's inventory).

### **Bounded Contexts**

We can separate the system into logical parts, each with its own model.

1. **Game World Context:** The core of the MUD. Manages the simulation, game rules, character state, and interactions. This is where most of our domain logic lives.  
2. **Authoring Context:** Tools and services for creating and managing game content (the DSL parsers, validators, and content pipeline).  
3. **Player Account Context:** Manages user accounts, authentication, and the list of characters associated with an account.  
4. **LLM Interaction Context:** A dedicated context for abstracting away the complexities of communicating with the Large Language Model.

### **Domain Model for the "Game World Context"**

This is the most critical context. We can model it using DDD patterns:

* **Aggregate Root: Character**  
  * This is the primary entity. An action is always performed by a Character. The Character aggregate is the consistency boundary for all its related data. When we save a character, we must save their inventory, stats, etc., in a single transaction.  
  * **Entities within Aggregate:**  
    * Inventory: A collection of Item instances.  
    * Stats: Manages attributes like health, mana, strength.  
  * **Value Objects:**  
    * Position: Holds the LocationID the character is currently in.  
    * HealthPoints: A simple value object for current/max health.  
* **Aggregate Root: Area**  
  * An outdoor location that manages its own state, connecting to other Areas via boundaries and containing Entrances to Structures.  
  * **Entities within Aggregate:**  
    * NpcInstance: A specific instance of an NPC template currently in this area.  
    * ItemInstance: An item on the ground in this area.  
  * **Value Objects:**  
    * Boundary: Defines a path to another AreaID.  
    * Entrance: A point of interest that leads to a StructureID.  
* **Aggregate Root: Structure**  
  * An enclosed location that manages its own state and connects to other Locations (either Areas or other Structures) via Exits.  
  * **Entities within Aggregate:**  
    * NpcInstance  
    * ItemInstance  
  * **Value Objects:**  
    * Exit: Defines a path to another LocationID.  
* **Repositories:**  
  * CharacterRepository: Handles loading and saving Character aggregates (e.g., from a database). FindByID(id), Save(character).  
  * WorldRepository: Handles loading the static game world definitions (Areas, Structures, Item Templates, NPC Templates) from the DSL files. GetLocation(id), GetItemTemplate(id).  
* **Services:**  
  * GameLoopService: The main ticker that processes time-based events.  
  * ActionService: Takes a parsed command from a player and executes it, coordinating between aggregates. For example, a MoveAction would involve the CharacterRepository (to load the character), the WorldRepository (to validate the boundary/exit from the current location), and the CharacterRepository again (to update the character's new position and save it).

### **Diagram of the Game World Model**

classDiagram  
    direction LR

    class Character {  
        \+UUID id  
        \+string name  
        \+Position position  
        \+Stats stats  
        \+Inventory inventory  
        \+Execute(Action)  
    }

    class Location {  
        \<\<interface\>\>  
        id  
        name  
        description  
        npcs  
        itemsOnGround  
    }

    class Area {  
        \+List\~Boundary\~ boundaries  
        \+List\~Entrance\~ entrances  
    }  
      
    class Structure {  
        \+List\~Exit\~ exits  
    }

    class Position {  
        \+string locationId  
    }  
    class Stats {  
        \+int health  
        \+int mana  
    }  
    class Inventory {  
        \+List\~Item\~ items  
    }  
    class Item {  
        \+UUID instanceId  
        \+string templateId  
        \+string name  
    }  
    class NPC {  
        \+UUID instanceId  
        \+string templateId  
        \+string name  
    }  
    class Boundary {  
        \+string direction  
        \+string targetAreaId  
    }  
    class Entrance {  
        \+string name  
        \+string targetStructureId  
    }  
    class Exit {  
        \+string direction  
        \+string targetLocationId  
    }

    Location \<|-- Area  
    Location \<|-- Structure

    Character "1" \-- "1" Position  
    Character "1" \-- "1" Stats  
    Character "1" \-- "1" Inventory  
    Inventory "1" \-- "0..\*" Item  
      
    Area "1" \-- "0..\*" Boundary  
    Area "1" \-- "0..\*" Entrance  
    Structure "1" \-- "0..\*" Exit  
      
    Location "1" \-- "0..\*" Item  
    Location "1" \-- "0..\*" NPC

## **Part 3: User Story Map**

A user story map helps organize development by mapping user activities to features, prioritized into releases.  
Backbone (User Activities / Epics):  
Account Management \-\> Character Management \-\> Core Gameplay Loop \-\> Social Interaction \-\> World Authoring

### **Release 1: Minimum Viable Product (The Core Engine)**

*Goal: A single player can log in, walk around a static world, and look at things.*  
*Developer Note: The implementation for these stories will now need to handle two distinct location types: Area and Structure, and the different ways to transition between them (Boundaries, Entrances, and Exits).*

| Activity | User Stories |
| :---- | :---- |
| **Account Mgmt** | As a user, I can register a new account. |
|  | As a user, I can log in to my account. |
| **Character Mgmt** | As a player, I can create a new character with a name. |
|  | As a player, I can select a character to play. |
| **Core Gameplay** | As a player, I can see the description of the room I am in (look). |
|  | As a player, I can move between rooms using exits (north, south, etc.). |
|  | As a player, I can see who else is in the room. |

### **Release 2: World Interaction**

*Goal: The world feels more alive with items and basic NPCs.*

| Activity | User Stories |
| :---- | :---- |
| **Character Mgmt** | As a player, I have an inventory to store items (inventory). |
| **Core Gameplay** | As a player, I can get an item from a room (get sword). |
|  | As a player, I can drop an item from my inventory (drop sword). |
|  | As a player, I can examine an item or NPC for more details (examine sword). |
|  | As a player, I can see static NPCs in rooms who have pre-scripted dialogue (talk to merchant). |

### **Release 3: The Dynamic World & Combat**

*Goal: Introduce game logic, conditions, and the combat loop.*

| Activity | User Stories |
| :---- | :---- |
| **Core Gameplay** | As a player, I can engage in simple turn-based combat with an NPC (attack goblin). |
|  | As a player, I can use items on other things or myself (use potion). |
|  | As a game author, I can define doors that require a specific key, using the expression language. |
|  | As a game author, I can create NPCs that react to player actions based on scripted conditions. |

### **Release 4: LLM Integration & Social Features**

*Goal: Make the world feel intelligent and allow players to communicate.*

| Activity | User Stories |
| :---- | :---- |
| **Core Gameplay** | As a player, I can have a dynamic, unscripted conversation with certain NPCs. |
|  | As a player, I can receive dynamic descriptions of rooms that reflect recent events. |
| **Social** | As a player, I can say things that everyone in the room can hear (say hello everyone). |
|  | As a player, I can whisper to another player in the same room (whisper john meet me by the docks). |

### **Future Releases: Authoring & Expansion**

* **World Authoring:**  
  * As a world builder, I can use a CLI tool to validate my DSL files for errors.  
  * As a world builder, I can hot-reload parts of the world on the running server without a restart.  
* **Protocol Expansion:**  
  * As a developer, I can define the gRPC services to stream room state to a client.  
  * As a visual client developer, I can connect to the gRPC endpoint to receive game updates.