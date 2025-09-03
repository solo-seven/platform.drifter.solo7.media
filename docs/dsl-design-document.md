# Design Document: Layered DSL for Role Playing Game Definition

## Overview

This design proposes a **layered domain-specific language (DSL)** for defining role playing games (RPGs). The goal is to support **pen-and-paper systems** as well as **computer-based RPG engines**, providing both human-friendly authoring and machine-friendly validation/serialization.

The DSL is structured into three complementary layers:

1. **Declarative Data Layer**
   Used for defining structured, referenceable game data (classes, items, monsters, encounters).

   * Format: **TOML** (preferred) or YAML.
   * Validated via JSON Schema for correctness.
   * Each entity has a stable `id` for cross-references.

2. **Expression Layer (Rules Kernel)**
   Embedded into the data layer for logic such as effects, conditions, dice rolls, and modifiers.

   * Syntax: **small deterministic expression language**.
   * Features: literals, operators, whitelisted functions (`roll("2d6")`, `heal()`, `deal()`, `has_tag()`).
   * Context-aware variables (`self`, `target`, `party`, `terrain`).
   * Deterministic, sandboxed, no loops/mutation.

3. **Prose/Presentation Layer**
   Used for lore, rulebooks, and player-facing narrative.

   * Format: **Markdown** with optional YAML/TOML front-matter.
   * Tagged inline references (`@Item/iron_sword`) link prose to game data.
   * Fenced blocks (` ```statblock `) render structured data inline.

## Example Workflow

1. **Authoring**: Designers write classes, monsters, and items in TOML, rules in expressions, and lore in Markdown.
2. **Validation**: Schema checks data, expression parser checks logic, CI enforces consistency.
3. **Compilation**: Build pipeline merges all layers into a normalized JSON bundle.
4. **Consumption**:

   * Pen-and-paper: Markdown compiles into PDFs/print rulebooks with embedded stat blocks.
   * Digital engines: Bundle provides normalized entities and rule logic for simulation.

## Example Snippets

### Fighter Class (TOML)

```toml
id = "class.fighter"
name = "Fighter"
hit_die = "d10"
primary_attributes = ["str", "con"]

[abilities.second_wind]
name = "Second Wind"
uses = { per = "short_rest", count = 1 }
effect = 'heal(self, roll("1d10") + self.level)'
```

### Iron Sword (TOML)

```toml
id = "item.iron_sword"
name = "Iron Sword"
type = "weapon"
damage = "1d8"
on_hit = 'deal(target, roll(damage), type="slashing")'
```

### Lore Document (Markdown)

````markdown
---
id: "doc.combat_overview"
title: "Combat Overview"
related: ["class.fighter", "item.iron_sword"]
---

Combat consists of a **Turn** with @Rules/Action, @Rules/Move, and @Rules/BonusAction.  
Fighters gain @Ability/second_wind at Level 1.

```statblock
ref: class.fighter
show: ["hit_die","primary_attributes"]
````

## Design Rationale
- **TOML for structure**: stricter, less error-prone than YAML, and easy to parse.  
- **Expressions for rules**: declarative and deterministic, bridging human-readable dice notation with engine execution.  
- **Markdown for lore**: widely adopted, good for prose, exportable to print/web, extensible with light tagging.  
- **Separation of concerns**: each layer optimizes for its audience (designers, GMs, engines, players).  

## Future Extensions
- Versioning (`dsl_version` field in each file).  
- Migration tools for schema evolution.  
- Optional higher-level GUI tools that generate/validate TOML + Markdown.  
- Export pipelines: JSON → VTT (Foundry, Roll20) or game engine integration.  

