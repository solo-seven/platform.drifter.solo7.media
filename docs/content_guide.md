# Game World for Drifters

This project is a a game world to allow the definition of the world through text and code. From there, the world is run
as an interactive simulation. From this, it will gather logs and information that can be used to create movies and
series from.

## Content Structure

### Game Master Guide

This contains the game system defintion which will define templates for actions that can be taken and the rules that
govern those actions.

#### Entity Templates

With Entity Templates, we define the combinations of components that will exist in the world and how they will interact
with each other. In order to define an entity, we need to define the components that exist.

### Player Guide

This defines the valid component combinations that the player can use for their avatar. All avatars will have a set of
core components assigned to them. One of those will be a skill manager that handles the skills which influence the
actions that the avatar can take.

Equipment templates will be defined within the player guide since it affects how the avatar can interact with the
world. Equipment instances will be defined within the world books as they are specific to the region they originate in.
That region must be attached to the world for the equipment to be available.

### Monster Manual

This provides the entity definitions that agents can use within the world. It also helps to define the behaviors that
agents can use to interact with the world.

### World Books

WOrld books are defined as regions. Within the world books are the possible entities and agents that will exist. It
also defines the physical map and structures that exist. The factions that exist in the region and how they affect the
world. Equipment catalogs exist in order to be assigned to physical locations within the region.