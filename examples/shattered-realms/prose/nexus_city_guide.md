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
