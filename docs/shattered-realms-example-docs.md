# The Shattered Realms: Complete DSL Example World

## 1. Game Master Definition (GMD)

### Core Mechanics Repository

```toml
# gmd/core_mechanics.toml
dsl_version = "1.0"
id = "gmd.shattered_realms"
name = "Shattered Realms Core Mechanics"
version = "1.2.0"

[fundamental_mechanics.reality_stability]
id = "mechanic.reality_stability"
name = "Reality Stability"
description = "Measures local coherence of physical laws"
base_value = 100
min_value = 0
max_value = 150
decay_rate = 'roll("1d6") - 3'
restoration_methods = ["rest", "ritual", "stabilizer_items"]

[fundamental_mechanics.cascade_failure]
id = "mechanic.cascade_failure"
name = "Cascade Failure"
description = "Chain reaction when reality drops below threshold"
trigger_condition = 'reality_stability <= 25'
effect = 'reality_stability -= roll("2d10"); trigger_environmental_hazard()'

[action_types.stabilize]
id = "action.stabilize_reality"
name = "Stabilize Reality"
action_cost = "full_action"
requires_skill = "arcana"
difficulty_class = 'max(10, 30 - reality_stability)'
effect = 'reality_stability += roll("1d8") + skill_modifier("arcana")'

[action_types.chaos_surge]
id = "action.chaos_surge"
name = "Chaos Surge"
action_cost = "bonus_action"
requires_attribute = "chaos_affinity"
risk_factor = 'reality_stability < 50 ? "high" : "moderate"'
effect = '''
  if (roll("1d20") + chaos_affinity > reality_stability) {
    deal(target, roll("2d6") + chaos_affinity, type="chaos");
    reality_stability -= roll("1d4")
  } else {
    deal(self, roll("1d6"), type="psychic");
    add_condition(self, "destabilized", duration="1_turn")
  }
'''
```

### Action Resolution System

```toml
# gmd/action_resolution.toml

[resolution_methods.stability_check]
id = "resolution.stability_check"
name = "Reality Stability Check"
applicable_actions = ["stabilize", "chaos_surge", "dimensional_travel"]
inputs = ["skill_modifier", "reality_stability", "environmental_modifiers"]
calculation = '''
  base_roll = roll("1d20") + skill_modifier;
  stability_bonus = reality_stability > 75 ? 2 : (reality_stability < 25 ? -3 : 0);
  total = base_roll + stability_bonus + environmental_modifiers;
  return total
'''

[conflict_resolution.magic_vs_stability]
priority = 100
scope = "magical_actions"
rule = '''
  if (action.type == "magical" && reality_stability < 50) {
    add_modifier(action, "unstable_magic", -2);
    chance_of_side_effect = (50 - reality_stability) / 10;
    if (roll("1d10") <= chance_of_side_effect) {
      trigger_wild_magic()
    }
  }
'''
```

## 2. Player's Guide (PG)

### Character Creation System

```toml
# pg/races.toml

[races.void_touched]
id = "race.void_touched"
name = "Void-Touched"
description = "Humans altered by exposure to reality fractures"
size = "medium"
speed = 30
lifespan = "80-120 years"

[races.void_touched.attributes]
charisma = 2
constitution = 1
chaos_affinity = 3  # Custom attribute

[races.void_touched.abilities]
void_sight = {
  id = "ability.void_sight",
  description = "Can see reality instabilities",
  effect = 'detect_reality_level(range=60)'
}

stability_drain = {
  id = "ability.stability_drain", 
  uses = { per = "long_rest", count = 1 },
  effect = 'target_location.reality_stability -= roll("1d6"); heal(self, roll("1d4"))'
}

[races.void_touched.resistances]
chaos = "advantage"
psychic = "resistance"
```

```toml
# pg/classes.toml

[classes.reality_warden]
id = "class.reality_warden"
name = "Reality Warden"
description = "Guardians who maintain the stability of existence"
hit_die = "d8"
primary_attributes = ["wisdom", "constitution"]
save_proficiencies = ["wisdom", "constitution"]

[classes.reality_warden.features.level_1.anchor_point]
name = "Anchor Point"
description = "Designate a location as a reality anchor"
uses = { per = "long_rest", count = 1 }
effect = '''
  anchor_location.reality_stability = max(anchor_location.reality_stability, 75);
  anchor_location.stability_decay = false;
  duration = "24_hours"
'''

[classes.reality_warden.features.level_3.detect_fractures]
name = "Detect Fractures"
uses = { per = "short_rest", count = 3 }
effect = 'reveal_all_fractures(range = 120); highlight_danger_zones()'

[classes.chaos_dancer]
id = "class.chaos_dancer"
name = "Chaos Dancer"
description = "Martial artists who channel instability"
hit_die = "d10" 
primary_attributes = ["dexterity", "chaos_affinity"]

[classes.chaos_dancer.features.level_1.unpredictable_strikes]
name = "Unpredictable Strikes"
description = "Attacks become more powerful in unstable reality"
effect = '''
  if (reality_stability < 50) {
    add_damage_dice(1);
    if (reality_stability < 25) {
      add_damage_dice(1);
      chance_teleport_after_attack = 0.3
    }
  }
'''
```

### Skill System

```toml
# pg/skills.toml

[skill_categories.reality_manipulation]
id = "category.reality_manipulation"
name = "Reality Manipulation"
description = "Skills for interacting with unstable reality"

[skills.stability_weaving]
id = "skill.stability_weaving"
name = "Stability Weaving"
category = "reality_manipulation"
governing_attributes = ["intelligence", "wisdom"]
description = "Art of reinforcing reality's structure"

[skills.stability_weaving.applications.emergency_stabilization]
context = "critical_failure"
difficulty = 'DC_15 + (50 - reality_stability) / 5'
success_effect = 'reality_stability += roll("2d4") + skill_modifier'
failure_effect = 'reality_stability -= roll("1d4"); add_condition(self, "disoriented")'

[skills.chaos_channeling]
id = "skill.chaos_channeling" 
name = "Chaos Channeling"
category = "reality_manipulation"
governing_attributes = ["charisma", "chaos_affinity"]

[skills.chaos_channeling.applications.controlled_surge]
context = "offensive_magic"
difficulty = 'DC_12 + (reality_stability > 75 ? 5 : 0)'
success_effect = '''
  damage = roll("1d8") + chaos_affinity;
  if (reality_stability < 40) damage += roll("1d4");
  deal(target, damage, type="chaos")
'''
```

### Equipment System

```toml
# pg/equipment.toml

[items.stabilizer_crystal]
id = "item.stabilizer_crystal"
name = "Stabilizer Crystal"
type = "wondrous_item"
rarity = "uncommon"
attunement_required = false

[items.stabilizer_crystal.properties]
reality_anchor = {
  effect = 'increase_local_stability(amount=10, radius=30)',
  duration = "permanent_while_held"
}

fragile = {
  condition = 'reality_stability <= 15',
  effect = 'item_destroyed(); reality_stability -= 5'
}

[items.chaos_blade]
id = "item.chaos_blade"
name = "Chaos Blade"  
type = "weapon"
weapon_type = "longsword"
damage = "1d8"
magical = true
attunement_required = true

[items.chaos_blade.properties]
variable_damage = {
  effect = '''
    base_damage = roll("1d8");
    if (reality_stability < 50) {
      extra_damage = roll("1d6");
      damage_type = random_choice(["fire", "cold", "lightning", "thunder"])
    }
    return base_damage + extra_damage
  '''
}

reality_disruption = {
  on_critical = 'target_location.reality_stability -= roll("1d4")'
}
```

## 3. Monster Manual (MM)

### NPC Definitions

```toml
# mm/creatures.toml

[npcs.fracture_wraith]
id = "npc.fracture_wraith"
name = "Fracture Wraith"
category = "undead_aberration"
challenge_rating = 5
threat_level = "moderate"

[npcs.fracture_wraith.attributes]
strength = 6
dexterity = 16  
constitution = 14
intelligence = 10
wisdom = 12
charisma = 15
chaos_affinity = 8

[npcs.fracture_wraith.abilities]
phase_through_reality = {
  description = "Move through solid matter in low-stability areas",
  condition = 'reality_stability < 40',
  movement_type = "incorporeal"
}

destabilize_touch = {
  attack_bonus = 7,
  damage = 'roll("2d6") + 3',
  damage_type = "necrotic",
  additional_effect = 'reality_stability -= roll("1d4")'
}

[npcs.fracture_wraith.behavior]
preferred_stability = "0-30"
aggression_modifier = 'reality_stability > 60 ? -2 : +3'
flee_condition = 'reality_stability > 90 || self.health < 0.25'

[npcs.stability_elemental]
id = "npc.stability_elemental"
name = "Stability Elemental"
category = "elemental_construct" 
challenge_rating = 6

[npcs.stability_elemental.attributes]
strength = 18
constitution = 20
wisdom = 14
chaos_affinity = 0  # Immune to chaos

[npcs.stability_elemental.abilities]
reality_reinforcement = {
  area_effect = "30_foot_radius",
  effect = 'reality_stability += roll("1d4") per turn (max 100)'
}

stability_beam = {
  range = 120,
  attack_bonus = 8,
  effect = '''
    if (target.chaos_affinity > 0) {
      deal(target, roll("3d8"), type="radiant");
      suppress_chaos_abilities(target, duration="1_minute")
    }
  '''
}
```

### Behavior Trees

```toml
# mm/behaviors.toml

[behavior_profiles.chaos_opportunist]
id = "profile.chaos_opportunist"
name = "Chaos Opportunist"
description = "Becomes more aggressive as reality destabilizes"

[behavior_profiles.chaos_opportunist.decision_tree]
root = "stability_assessment"

[behavior_profiles.chaos_opportunist.nodes.stability_assessment]
type = "condition"
condition = 'reality_stability < 50'
if_true = "aggressive_tactics"
if_false = "cautious_approach"

[behavior_profiles.chaos_opportunist.nodes.aggressive_tactics]
type = "sequence"
actions = [
  "move_to_optimal_position",
  "use_chaos_abilities", 
  "destabilize_environment"
]

[behavior_profiles.chaos_opportunist.nodes.cautious_approach]
type = "selector"
options = [
  "maintain_distance",
  "seek_cover",
  "prepare_defensive_abilities"
]
```

### Encounter Generation

```toml
# mm/encounters.toml

[encounter_types.reality_storm]
id = "encounter.reality_storm"
name = "Reality Storm"
description = "Chaotic weather patterns that warp local physics"
environment_types = ["any"]
stability_requirements = "10-40"

[encounter_types.reality_storm.phases]
warning = {
  duration = "2_rounds",
  effects = ['reality_stability -= 5 per round', 'add_condition("disoriented", all_creatures)']
}

peak_storm = {
  duration = "3_rounds", 
  effects = [
    'reality_stability -= roll("2d6") per round',
    'random_teleportation(chance=0.2, distance="30_feet")',
    'wild_magic_surge(chance=0.4)'
  ]
}

[encounter_types.stability_breach]
id = "encounter.stability_breach"
name = "Stability Breach"
creature_types = ["fracture_wraith", "chaos_spawn"]
composition_rules = [
  { creature = "fracture_wraith", count = "1d4" },
  { creature = "chaos_spawn", count = "2d6", condition = "party_level >= 5" }
]
environmental_effects = [
  'reality_stability starts at roll("2d20") + 20',
  'decreases by roll("1d4") each round'
]
```

## 4. World Books (WB)

### Location Definitions

```toml
# wb/regions/shattered_coast.toml

[world_book]
id = "wb.shattered_coast"
title = "The Shattered Coast"
description = "A coastal region where reality fractures meet the sea"
version = "1.0"

[locations.nexus_city]
id = "location.nexus_city"
name = "Nexus City"
type = "major_settlement"
population = 45000
base_reality_stability = 85

[locations.nexus_city.districts.stable_quarter]
id = "district.stable_quarter"
name = "The Stable Quarter"
reality_stability = 95
description = "Wealthy district maintained by Stability Guilds"
special_features = [
  "Reality Anchor Network",
  "Chaos Detection Wards", 
  "Emergency Stabilization Centers"
]

[locations.nexus_city.districts.flux_bazaar]
id = "district.flux_bazaar" 
name = "Flux Bazaar"
reality_stability = 60
description = "Market district where chaos-touched goods are traded"
random_events = [
  { chance = 0.1, event = "merchant_spontaneous_teleport" },
  { chance = 0.05, event = "goods_temporarily_phase_out" },
  { chance = 0.15, event = "currency_changes_metal_type" }
]

[locations.fracture_wastes]
id = "location.fracture_wastes"
name = "The Fracture Wastes"
type = "dangerous_wilderness"
base_reality_stability = 25
stability_variance = "1d20"

[locations.fracture_wastes.hazards]
floating_islands = {
  frequency = "common",
  effect = 'navigation_difficulty += 5; chance_of_fall_damage = 0.3'
}

time_loops = {
  frequency = "rare",
  trigger = 'reality_stability <= 15',
  effect = 'repeat_last_combat_round(); reality_stability -= 10'
}

gravity_wells = {
  frequency = "uncommon", 
  area = "50_foot_radius",
  effect = 'difficult_terrain; strength_save_or_pulled_toward_center'
}
```

### Environmental Systems

```toml
# wb/environmental/reality_weather.toml

[environmental_systems.reality_weather]
id = "system.reality_weather"
name = "Reality Weather Patterns"
affected_locations = ["all_outdoor"]
system_type = "metaphysical_weather"

[environmental_systems.reality_weather.patterns.stability_storm]
name = "Stability Storm"
frequency = "weekly"
duration = "2d6_hours"
conditions = [
  'reality_stability increases by 2 per hour',
  'chaos_abilities suffer disadvantage', 
  'stabilizer_items recharge faster'
]
side_effects = [
  'magic_items_glow_brightly',
  'chaotic_creatures_seek_shelter'
]

[environmental_systems.reality_weather.patterns.chaos_winds]
name = "Chaos Winds"
frequency = "bi_weekly"
duration = "1d4_hours"
intensity_levels = ["light", "moderate", "severe"]

[environmental_systems.reality_weather.patterns.chaos_winds.effects.light]
reality_stability_modifier = -1
spell_failure_chance = 0.05
random_color_changes = true

[environmental_systems.reality_weather.patterns.chaos_winds.effects.severe] 
reality_stability_modifier = -5
spell_failure_chance = 0.25
effects = [
  'random_polymorph(duration="10_minutes", chance=0.1)',
  'temporary_ability_score_swap',
  'spoken_words_become_visible'
]
```

### Cultural Systems

```toml
# wb/cultures/stability_guilds.toml

[cultures.stability_guild]
id = "culture.stability_guild"
name = "Stability Guilds"
dominant_locations = ["nexus_city", "anchor_towns"]
population_percentage = 0.15

[cultures.stability_guild.social_hierarchy]
guild_master = { authority = 100, requirements = ["stability_weaving >= 18", "leadership >= 15"] }
senior_weavers = { authority = 70, requirements = ["stability_weaving >= 12"] }
apprentices = { authority = 20, requirements = ["stability_weaving >= 5"] }

[cultures.stability_guild.economic_systems]
primary_trade = "stability_services"
monopolies = ["stabilizer_crystals", "reality_anchoring"] 
trade_goods = [
  { item = "stabilizer_crystal", price_modifier = 1.0 },
  { item = "chaos_dampener", price_modifier = 0.8 },
  { item = "reality_anchor", price_modifier = 2.0 }
]

[cultures.stability_guild.values]
order = { priority = 1, manifestation = "strict_guild_hierarchies" }
predictability = { priority = 2, manifestation = "standardized_procedures" }  
preservation = { priority = 3, manifestation = "reality_conservation_efforts" }

[cultures.chaos_dancers]
id = "culture.chaos_dancers"
name = "Chaos Dancer Clans"
social_structure = "fluid_tribal"
dominant_locations = ["flux_bazaar", "fracture_wastes"]

[cultures.chaos_dancers.traditions]
coming_of_age = {
  name = "First Surge",
  requirement = "survive_chaos_storm_solo",
  reward = "chaos_affinity +1, clan_tattoo"
}

seasonal_gathering = {
  name = "Convergence Festival", 
  frequency = "annual",
  location = "wherever_reality_is_weakest",
  activities = ["chaos_dancing", "instability_contests", "mutation_celebration"]
}
```

## 5. Integration Examples

### Cross-Reference Validation

```toml
# validation/cross_references.toml

[validation_rules]
# Ensure all referenced abilities exist
ability_references = {
  pattern = 'ability\.[a-z_]+',
  must_exist_in = "abilities/*.toml"
}

# Verify stability thresholds are consistent
stability_thresholds = {
  chaos_surge_risk = { min = 0, max = 50 },
  fracture_wraith_preference = { min = 0, max = 30 },
  stability_storm_trigger = { min = 75, max = 100 }
}

# Check that all locations have valid stability ranges
location_stability = {
  all_locations_must_have = "base_reality_stability",
  valid_range = { min = 0, max = 150 }
}
```

### Content Dependencies

```toml
# dependencies/module_graph.toml

[content_modules.core_mechanics]
exports = ["reality_stability", "chaos_affinity", "stability_actions"]
version = "1.2.0"

[content_modules.shattered_coast_region]
imports = ["core_mechanics.reality_stability", "core_mechanics.chaos_affinity"] 
requires = { core_mechanics = ">=1.0.0" }
exports = ["nexus_city", "fracture_wastes", "reality_weather"]

[content_modules.chaos_dancer_culture]
imports = [
  "core_mechanics.chaos_affinity",
  "shattered_coast_region.flux_bazaar"
]
requires = { 
  core_mechanics = ">=1.2.0",
  shattered_coast_region = ">=1.0.0" 
}
```

## 6. Prose/Presentation Layer Examples

### World Book Chapter (Markdown)

````markdown
---
id: "doc.nexus_city_guide"
title: "Nexus City: A Visitor's Guide"
related: ["location.nexus_city", "culture.stability_guild"]
dsl_version: "1.0"
---

# Nexus City: Where Order Meets Chaos

Nexus City stands as a beacon of stability in the @Region/shattered_coast, its towering spires anchored in reality by the @Organization/stability_guilds. With a @Stat/base_reality_stability of 85, it's one of the most reliable locations in the known world.

## Districts and Quarters

### The Stable Quarter
The crown jewel of Nexus City, the @Location/stable_quarter maintains near-perfect reality coherence through an intricate network of @Item/reality_anchor crystals. Here, the wealthy merchants and @NPC/guild_master officials conduct business with the certainty that their gold will remain gold and their buildings will stay properly attached to the ground.

```statblock
ref: location.nexus_city.districts.stable_quarter
show: ["reality_stability", "special_features"]
```

### Flux Bazaar
A more... adventurous district, the @Location/flux_bazaar caters to those who trade in @Item/chaos_touched_goods. Visitors should be prepared for the occasional merchant who phases in and out of reality, or currency that spontaneously transmutes.

**Survival Tip**: Always count your change twice in the Flux Bazaar—once before the transaction, and once after reality settles.

## Notable Inhabitants

The @Class/reality_warden patrol the city's borders, their @Ability/anchor_point abilities creating stable zones for travelers. Meanwhile, @Race/void_touched refugees from the outer fractures bring tales of the impossible.

```encounter
type: stability_breach
location: flux_bazaar
trigger: "reality_stability drops below 40"
description: "Reality cracks appear in the cobblestones, and @NPC/fracture_wraith begin seeping through."
```

## Visiting the City

- **Best Time**: During @Weather/stability_storm season for guaranteed coherent architecture
- **Avoid**: @Weather/chaos_winds—they make navigation difficult and may cause temporary polymorph
- **Bring**: @Item/stabilizer_crystal for emergencies, especially in the outer districts
- **Don't Miss**: The weekly @Event/convergence_festival in the Flux Bazaar (when reality permits)
````

### Player Handout (Markdown)

````markdown
---
id: "handout.chaos_dancer_training"
title: "Chaos Dancer Training Manual"
audience: "player"
related: ["class.chaos_dancer", "skill.chaos_channeling"]
---

# Embrace the Unpredictable: Chaos Dancer Fundamentals

*As passed down through the @Culture/chaos_dancer_clans*

## Core Philosophy
Where others fear the @Mechanic/reality_stability fluctuations, we dance with them. Each crack in reality is an invitation, each surge of chaos a partner in the eternal dance.

## Basic Techniques

### Unpredictable Strikes
Your @Ability/unpredictable_strikes become more potent as reality weakens around you:

```statblock
ref: class.chaos_dancer.features.level_1.unpredictable_strikes
show: ["effect", "stability_thresholds"]
```

When reality drops below 50, you gain additional power. When it falls below 25, you may find yourself teleporting unexpectedly after strikes—use this to your advantage.

### Reading the Flux
Develop your @Skill/chaos_channeling by practicing these exercises:

1. **Meditation in Storms**: Sit quietly during @Weather/chaos_winds and feel the reality fluctuations
2. **Crystal Resonance**: Hold a @Item/chaos_touched crystal and learn to sense its instability
3. **Partner Practice**: Spar near @Location/fracture_sites to experience variable physics

## Advanced Training

Once you've mastered the basics, seek out the @Location/fracture_wastes for true testing. There, with @NPC/fracture_wraith as opponents and shifting gravity as your training ground, you'll learn to make chaos your ally.

Remember: *"In stability we stagnate, in chaos we grow."* —First Dancer Kaelith the Ever-Shifting
````

This example demonstrates how all layers of the DSL work together to create a cohesive game world with:

1. **Technical depth** in the data structures and rule definitions
2. **Mechanical innovation** with the reality stability system
3. **Narrative coherence** linking all components through consistent themes
4. **Cross-references** that enable rich content relationships
5. **Validation frameworks** that ensure consistency across modules

The reality stability mechanic serves as a unifying thread, affecting combat, exploration, character abilities, environmental systems, and cultural dynamics throughout the world.
