# The Shattered Realms: Complete DSL Example World

This directory contains a complete implementation of the Shattered Realms example world as defined in the DSL design document. This serves as a comprehensive testbed for implementing and validating our DSL system.

## Directory Structure

```
examples/shattered-realms/
├── gmd/                    # Game Master Definition
│   ├── core_mechanics.toml
│   └── action_resolution.toml
├── pg/                     # Player's Guide
│   ├── races.toml
│   ├── classes.toml
│   ├── skills.toml
│   └── equipment.toml
├── mm/                     # Monster Manual
│   ├── creatures.toml
│   ├── behaviors.toml
│   └── encounters.toml
├── wb/                     # World Books
│   ├── regions/
│   │   └── shattered_coast.toml
│   ├── environmental/
│   │   └── reality_weather.toml
│   └── cultures/
│       └── stability_guilds.toml
├── validation/
│   └── cross_references.toml
├── dependencies/
│   └── module_graph.toml
├── prose/
│   ├── nexus_city_guide.md
│   └── chaos_dancer_training.md
└── README.md
```

## Key Features Demonstrated

### Core Mechanics
- **Reality Stability System**: A unified mechanic that affects all aspects of gameplay
- **Cascade Failure**: Chain reactions when reality drops below critical thresholds
- **Action Resolution**: Complex resolution methods with environmental modifiers

### Character System
- **Custom Attributes**: `chaos_affinity` as a new core attribute
- **Reality-Based Abilities**: Powers that interact with the stability system
- **Dynamic Class Features**: Abilities that scale with environmental conditions

### World Building
- **Environmental Systems**: Weather patterns that affect reality stability
- **Cultural Integration**: Societies built around the core mechanics
- **Location-Based Mechanics**: Different areas with varying stability levels

### Cross-References
- **Validation Rules**: Ensuring consistency across all content
- **Module Dependencies**: Managing content relationships and versions
- **Prose Integration**: Markdown documents that reference game elements

## Usage for DSL Implementation

This example world provides:

1. **Comprehensive Test Cases**: Every major DSL feature is demonstrated
2. **Complex Relationships**: Cross-references between different content types
3. **Real-World Complexity**: A complete, playable game world
4. **Validation Scenarios**: Examples of what needs to be validated
5. **Integration Examples**: How different systems work together

## Next Steps

With this example world implemented, we can now:

1. **Parse and Validate**: Test our TOML parsing and validation systems
2. **Resolve References**: Implement cross-reference resolution
3. **Generate Content**: Create tools to process and present the content
4. **Test Gameplay**: Use this as a testbed for the game engine
5. **Iterate on Design**: Refine the DSL based on real usage

This example demonstrates the full power and flexibility of our DSL system, showing how it can support complex, interconnected game worlds with rich mechanical depth.
