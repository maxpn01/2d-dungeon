# V2 â€” Second Floor: Forest

## Goal
- v2 is complete when a run can go floor 1 â†’ forest floor 2 â†’ defeat the cave hydra â†’ teleport to floor 3.
- Reaching floor 3 wins the run. Floor 3 is only a small room for now, in a volcanic magma / netherworld style.
- The game is a roguelike, similar in spirit to Magicraft: one run, permanent death, power gained during the run.
- The forest should feel drastically different from floor 1 through layout, enemies and rewards, while keeping the existing movement, combat, eating and growth mechanics.

## Decisions
These are settled. Numbers marked *(tune)* are starting values to balance during play.

### Progression
- Taking the descent portal on floor 1 no longer ends the run and shows no results screen. The player just arrives on floor 2.
- The return portal on floor 2 still works, so the player can go back to floor 1.
- The results / win screen moves to the moment the player reaches floor 3.
- Dying anywhere, including floor 2, restarts the whole run.
- Clearing floor 1 still plays the completion sound and effect, and shows a "FLOOR I COMPLETE" popup for a few seconds (`parameters.FloorBannerSeconds`).
- Reaching floor 3 plays the same completion sound and effect, shows a "FLOOR II COMPLETE" popup, and opens the results screen with the same title.
- Until the hydra exists, floor 2 has a placeholder forest-exit portal to floor 3 that is always active. Milestone 8 replaces it with the hydra's teleport circle.

### Forest map
- Replace the current floor 2 room with a much larger forest of winding paths, dense trees and clearings.
- Build the forest with a new layout method (for example a hand-drawn tile/ASCII map) instead of rectangular rooms joined by straight passages. Clearings can still be rooms, so patches, statues and the boss-arena role keep working.
- Paths branch and reconnect into loops, giving room to explore, approach from different sides and escape from large snakes.
- Landmarks, food patches and perk statues sit along the routes so the extra space has a purpose.
- Several routes eventually lead to a hidden cave with the boss. Shed skins and large tracks hint at where it is.
- Keep the current food. No forest-specific food types or visuals.
- The forest is 200×200 cells, placed east of floor 1 (origin `parameters.ForestOriginX/Y`). It is drawn by hand in `internal/worldmap/forest.txt`, which is the source of truth and is loaded with `go:embed`.
- Map legend: `.` grass, `,` dirt path (drawn as a bridge where it crosses water), `"` bush, `T` tree, `F` fruit tree, `~` water, `#` rock, `:` cave floor, `O` giant stump, `L` fallen log, `R` ruin wall, `M` standing stone. Markers on ground: `P` return portal, `X` forest-exit portal, `S` perk statue, `f` feeding spot.
- Trees, fruit trees, water and landmarks block movement but not sight. Trees and landmarks also stop projectiles; shots fly over water. Bushes are decoration only. Rock is the only terrain that blocks sight, which keeps the cave hidden.
- Fruit trees drop a little of the current food around themselves on a timer, as small snacks between the larger feeding spots.
- Floor 2 has 4 perk statues, one per quarter of the forest. They offer floor-1 perks until the forest pool exists (milestone 7).
- The hidden cave sits in the north-east corner, ringed by rock, with three tunnels that connect to different forest routes. The placeholder forest-exit portal is inside the cave.
- AI pathing uses A* over the cell grid so it scales to the forest's size.
- Floor 2 has its own renderer: a tilted low-poly 3D view in the style of Polytopia, in `internal/rendering/worldmap/forest3d*.go`. Floor 1 keeps the flat 2D renderer. Gameplay stays 2D: actors, food, projectiles and effects use their existing 2D drawing, rendered into a texture laid on the ground plane, so trees in front hide them. Canopies dither away near the player, and mouse aim ray-casts onto the ground.
- The forest mood is enchanted twilight: a warm low sun, teal ambient light, a lantern glow around the player, glowing mushrooms, fruit, crystals and runes, fireflies and light shafts in clearings.
- In the forest, fog of war dims unexplored and remembered terrain instead of hiding it; enemies and items are still hidden outside vision.

### Patrolling snakes
- Green snakes start with 3–4 body segments (random) and patrol routes between feeding spots.
- They eat the food at those spots, competing with the player, and grow longer and stronger.
- Each growth step needs a bit more food than the last; the increase should stay modest. Health and length grow more freely than damage and speed.
- Maximum length: 20 segments. A snake at full length stops growing entirely: no more length and no more stats.
- Like other mobs, snakes are not solid: the player passes through them. Body segments deal no contact damage; only the head's attacks hurt.
- Snakes respawn at their starting size after about 1 minute, the same as floor 1 mobs and perk statues.
- The player can kill snakes early, take the food before they arrive, or leave them to grow into bigger threats.
- Use the archived serpent boss (`_archive/prototype/internal/simulation/boss*.go`) as a starting reference for body following, feeding and attacks.
- Snakes are generated from the map, not placed by hand: 20 snakes (`parameters.ForestSnakeCount`) are spread over the forest feeding spots outside the cave, so some spots are home to two. A snake's patrol route is its home spot plus the 2 nearest other feeding spots, visited in a loop with a short pause at each; a second snake at the same spot walks the route in reverse, starting from the far end.
- Snakes start at level 4 with 2.5× the normal mob health *(tune)*, so they survive long enough to grow. They slither slower on patrol than when chasing.
- A snake chases the player it can see within a short radius, as long as the player stays near one of the snake's route stops (`parameters.SnakeChaseLeash`). After losing the player it searches, then returns home.
- Snakes are slippery instead of relentless: after each bite they back off (to about 170 px, for under a second) while still facing the player, then weave back in with a side-to-side sway that makes them harder to hit. Taking a burst of damage (20% of max HP in a short window) also makes them back off.
- Body segments follow the head and can be hit by projectiles; damage goes to the snake's single HP pool.
- Snakes are drawn as low-poly 3D actors in the forest view: a head with eyes and a tapered, banded body.

### Snake attacks
- **Bite:** close-range snap that deals damage and applies poison. Short head pullback and a cooldown.
- **Spit:** a clearly telegraphed poison projectile.
- **Fast slither:** the snake briefly coils, then surges forward with limited turning, so the player can dodge out of its path.

### Status effects and poison
- Add a shared status-effect system that works on players and mobs. Poison is the first effect.
- Poison deals damage at regular intervals for a limited time, but never drops the target below 1 HP.
- Poison strength scales with the attacker's size and strength. Duration stays fairly constant at first *(tune)*.
- When poison is reapplied it refreshes the duration instead of stacking. A stronger poison replaces a weaker one; a weaker hit never downgrades it.
- Show it clearly: the health bar turns purple, plus a poison icon with the remaining time, and readable damage ticks.
- In god mode and during invulnerability windows the effect is still shown but deals no damage.

### Perks
- Add a forest perk pool. Like the existing perks, each forest perk has ranks Eâ€“S.
- Floor 2 statues offer a mix of first-floor and forest perks. If the player already has every first-floor perk at max rank, they offer only forest perks.
- Forest perks change mechanics and use the existing controls. Starting candidates *(tune)*:
  - **Venom:** hits apply a short poison; repeated hits refresh it.
  - **Pierce:** a projectile passes through extra enemies with reduced damage. One shot damages each snake at most once.
  - **Shedskin:** periodically clears poison and briefly blocks it from being reapplied.
  - **Hunter's Instinct:** bonus damage against enemies above a high HP threshold (for example 70%), rewarding early attacks on growing snakes.

### Cave boss: three-headed hydra
- One shared HP bar. Both the heads and the body can be damaged.
- When a head is killed, the hydra can no longer attack with it, only with the remaining heads.
- A destroyed head regrows after about 1 minute *(tune)*. Regrowing costs mana, so the hydra needs enough mana saved up.
- Attacks: staggered poison spit from the heads; only one head lunges at a time.
- The hydra has its own level track, starting at level 1.
- It grows by eating food that spawns in the cave and by killing snakes. Killing a snake gives it XP, HP and MP the same way kills do for other actors. Growing makes it bigger and stronger, but never adds heads or segments.
- It follows a repeating cycle *(tune)*:
  1. Sleeps for 2 minutes.
  2. Looks for food inside the cave for 1 minute.
  3. Hunts outside the cave for 1 minute.
  4. Returns to the cave and sleeps again.
- While hunting it picks targets carefully: only snakes it can safely beat. Snakes can still kill it in principle, but it should almost never happen.
- Snakes killed by the hydra respawn normally.
- The hydra does not respawn. Once killed, it stays dead for the rest of the run.
- Defeating it activates the teleport circle to floor 3.

## Milestones
Build in this order. Each milestone should be playable on its own.

1. **Progression** *(done)*: the descent portal leads into floor 2 without ending the run; the win screen moves to a floor 3 placeholder.
2. **Forest map** *(in playtest)*: the new layout method, the forest itself, and navigation (minimap, fog of war, AI pathing).
3. **Basic snake** *(in playtest)*: segmented body, patrol routes, bite attack, respawn.
4. **Status effects:** the shared system, poison, and its HUD and visuals.
5. **Feeding and growth:** snakes eat, grow with rising costs, 20-segment cap.
6. **More snake attacks:** spit and fast slither.
7. **Forest perks:** the forest pool, ranks and the mixed offers at floor 2 statues.
8. **Cave and hydra:** cave clues, the cave arena, the hydra fight, head regrowth, feeding and hunting.
9. **Floor 3:** a small volcanic room reached through the hydra's teleport circle, ending the run with the win screen.

## Out of scope for v2
- Forest-specific food.
- Floor 3 content beyond a single small room.
- Progression that carries over between runs (meta-progression).

## Open questions
- Exact numbers, to find by playtesting: poison damage and tick rate, snake growth costs, perk values for each rank, the hydra's levels and cycle timings.
