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
- The forest-exit portal to floor 3 sits in the hydra's lair inside the cave. It stays sealed ("SEALED BY THE HYDRA'S BLOOD") until the hydra is dead.

### Forest map
- Replace the current floor 2 room with a much larger forest of winding paths, dense trees and clearings.
- Build the forest with a new layout method (for example a hand-drawn tile/ASCII map) instead of rectangular rooms joined by straight passages. Clearings can still be rooms, so patches, statues and the boss-arena role keep working.
- Paths branch and reconnect into loops, giving room to explore, approach from different sides and escape from large snakes.
- Landmarks, food patches and perk statues sit along the routes so the extra space has a purpose.
- Several routes eventually lead to a hidden cave with the boss. Shed skins and large tracks hint at where it is.
- Keep the current food. No forest-specific food types or visuals.
- The forest is 200×200 cells, placed east of floor 1 (origin `parameters.ForestOriginX/Y`). It is drawn by hand in `internal/worldmap/forest.txt`, which is the source of truth and is loaded with `go:embed`.
- Map legend: `.` grass, `,` dirt path (drawn as a bridge where it crosses water), `"` bush, `T` tree, `F` fruit tree, `~` water, `#` rock, `:` cave floor, `O` giant stump, `L` fallen log, `R` ruin wall, `M` standing stone. Markers on ground: `P` return portal, `X` forest-exit portal and hydra lair, `S` perk statue, `f` feeding spot (a cave feeding spot when it sits on cave floor), `s` shed skin, `h` hydra tracks. A marker's ground is cave floor next to the cave, dirt path when at least two neighbours are path, and grass otherwise.
- Trees, fruit trees, water and landmarks block movement but not sight. Trees and landmarks also stop projectiles; shots fly over water. Bushes are decoration only. Rock is the only terrain that blocks sight, which keeps the cave hidden.
- Fruit trees drop a little of the current food around themselves on a timer, as small snacks between the larger feeding spots.
- Floor 2 has 4 perk statues, one per quarter of the forest. They offer floor-1 perks until the forest pool exists (milestone 7).
- The hidden cave sits in the north-east corner, ringed by rock, with three tunnels that connect to different forest routes. The forest-exit portal is inside the cave, at the hydra's lair.
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
- In the forest view, hit and poison damage numbers are drawn in screen space above the actor (`parameters.DamageNumberLift`), so the 3D bodies and trees never hide them.
- Snakes eat with their head only, any food the player can eat (patch food, fruit-tree drops and kill drops). While patrolling or pausing at a stop, a snake slithers to the nearest food within `parameters.SnakeForageRadius` (170 px) and gives up on a piece it cannot reach after `SnakeForageTimeoutTicks` (6 s). Chasing the player takes priority over feeding. Snakes only forage and eat while the player is on floor 2, so they don't grow while the player is still on floor 1.
- Snakes grow through levels. Each piece of food heals, restores mana and gives `parameters.SnakeFoodXP` (20) XP per point of food value (its HP, or mana converted at the `FruitMP`:`FruitHP` rate) *(tune)*. Each level gained adds one tail segment, so a snake's level shows how much it has grown. The level costs follow the normal XP curve, so each step needs a bit more food than the last, and stats rise at the normal per-level rates. Current HP scales with the new max HP.
- At `SnakeMaxSegments` (20) a snake stops gaining XP and levels, though it still eats to heal and keeps the food from the player.
- A grown snake drops loot for its grown max HP. A respawned snake starts again at its respawn level with 3–4 segments.

### Snake attacks
- **Bite:** close-range snap that deals damage and applies poison. Short head pullback and a cooldown.
- **Spit:** a clearly telegraphed poison projectile.
- **Fast slither:** the snake briefly coils, then surges forward with limited turning, so the player can dodge out of its path.
- Spit and slither are used while chasing, outside bite range and only with a clear lane: spit at 130-340 px when a shot can reach the player, slither at 140-300 px when the snake can move straight there. If both are ready the snake picks one at random. Each has its own cooldown (spit 4 s, slither 6 s, plus up to 1 s jitter), and a 1.5 s gap follows any special attack *(tune)*.
- Spit: the head rears back for 0.6 s while a venom glob grows in its mouth and a dashed aim line points at the player; the aim tracks the player (with lead) until 0.15 s before release, when the line turns green. The glob flies at 520 px/s, deals 0.7× the snake's damage and poisons for `SnakeSpitPoisonSeconds` (3 s), using the same poison strength as the bite. Costs 1 MP. Blocked by trees and landmarks like other shots *(tune)*.
- Fast slither: the snake coils for 0.65 s with a shaking head and an orange lane with moving chevrons showing the charge; it aims freely until 0.2 s before the surge. The surge lasts 0.5 s at 2.8× its speed, turning at most 1.4 rad/s toward the player. Contact deals 1.2× damage plus bite poison, once per surge. The surge ends on a hit or when it runs into terrain, then the snake recovers for 0.45 s and backs off like after a bite *(tune)*.
- Snakes commit to spit and slither: taking damage does not cancel them.

### Status effects and poison
- Add a shared status-effect system that works on players and mobs. Poison is the first effect.
- Poison deals damage at regular intervals for a limited time, but never drops the target below 1 HP.
- Poison strength scales with the attacker's size and strength. Duration stays fairly constant at first *(tune)*.
- When poison is reapplied it refreshes the duration instead of stacking. A stronger poison replaces a weaker one; a weaker hit never downgrades it.
- Show it clearly: the health bar turns purple, plus a poison icon with the remaining time, and readable damage ticks.
- In god mode and during invulnerability windows the effect is still shown but deals no damage.
- Effects live in `simulation.StatusEffects`, carried by both `Player` and `AI` and advanced once per tick in `World.Step`. They clear when the carrier dies.
- Poison ticks every `parameters.PoisonTickSeconds` (0.5 s). Damage per tick is the attacker's damage × `PoisonDamageRatio` (0.15) × (1 + `PoisonSegmentBonus` (0.04) per body segment) *(tune)*. A snake bite poisons for `SnakePoisonSeconds` (4 s) *(tune)*.
- Reapplying keeps the stronger damage and the longer remaining time, without resetting the tick timer.
- Visuals: the HUD HP bar turns purple with a poison icon, a countdown ring and the remaining seconds; the screen edge pulses purple; poisoned actors pulse purple and give off rising bubbles; poison ticks show as small purple numbers without the hurt flash, shake or sound.

### Perks
- Add a forest perk pool. Like the existing perks, each forest perk has ranks E–S.
- Floor 2 statues offer a mix of first-floor and forest perks. If the player already has every first-floor perk at max rank, they offer only forest perks.
- Forest perks change mechanics and use the existing controls:
  - **Venom:** hits poison the target for `VenomPoisonSeconds` (2 s); each poison tick deals 8/10/12/15/18/24% of the shot's damage *(tune)*. Repeated hits refresh it.
  - **Pierce:** a shot passes through 1/1/2/2/3/4 extra enemies, keeping 45/55/60/65/70/80% of its damage (and venom) after each hit *(tune)*. One shot damages each enemy at most once, so a snake's head and body count as one target. Walls and trees still stop it.
  - **Shedskin:** when poisoned, the player sheds it after `ShedskinDelaySeconds` (0.35 s) and is immune to poison for 1/1.25/1.5/2/2.5/3 s, then it recharges for 20/17/14/12/10/8 s *(tune)*. The HUD shows its state, and a shed ring bursts around the player.
  - **Hunter's Instinct:** +15/25/35/50/65/85% damage against targets at or above `HunterHPThreshold` (70%) HP *(tune)*.
- Forest perks are `PerkFamily` values from `FirstForestPerk` on, so they share the perk arrays (`parameters.PerkBonuses`), ranks, cards and stats panel with floor-1 perks.
- A forest statue's offer always holds at least one forest card while one can still be upgraded; the rest are drawn from both pools. Floor-1 statues and mob perk milestones use only the floor-1 pool.
- Spectral echo shots carry the same forest perk effects.

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

### Hydra implementation
- The hydra is a mob of type `CaveHydra`, spawned from a den at the `X` marker (`worldmap.HydraLair`). Its code lives in `simulation/hydra.go` (cycle, targeting, movement) and `simulation/hydra_heads.go` (heads, attacks, severing, regrowth). Balance numbers are the `Hydra*` values in `parameters.go`.
- Level track: max HP is `HydraBaseHealth` (1100) plus `HydraHealthPerLevel` (110) per level, and damage is `HydraBaseDamage` (4) plus `HydraDamagePerLevel` (.4) per level. Each level also grows it by 4%, up to 1.6× size. Level costs follow the normal XP curve, and current HP scales with max HP *(tune)*.
- Each head has its own HP: `HydraHeadHealth` (20%) of max HP. A hit on a head damages both that head and the shared pool. Hits on the body or tail, and poison, only damage the pool. One shot damages the hydra at most once, so pierce counts it as one target.
- A severed head leaves a bleeding stump. After `HydraRegrowSeconds` (60 s) it regrows over 1.5 s, once the hydra has `HydraRegrowCost` (6) MP. While a regrowth is due within 12 s, the hydra holds back enough mana for it instead of spitting. With no heads left it cannot attack and backs away.
- Spit volley: each free head rears back for 0.6 s with a dashed aim line, one after another, 0.3 s apart. Each aims with lead and a small spread. Globs fly at 560 px/s, deal 1× damage, poison for 3 s and cost 0.5 MP each. Volleys come every 2.2 s at 110-520 px when the lane is clear *(tune)*.
- Lunge: one head at a time pulls back for 0.55 s under an orange lane, which turns red 0.15 s before the strike. Then it snaps out to 210 px for 2.2× damage. The body holds still during a lunge *(tune)*.
- In a fight the hydra keeps about 190 px from its target, circles slowly, and backs off when closer than 110 px. It turns toward its target at a limited rate.
- Sleeping: it lies coiled on the sealed circle with its eyes closed and heals 1% of max HP and 0.5 MP per second. It wakes when hurt or when the player comes within 220 px in sight, roars, and starts foraging.
- Awake, it attacks a player it can see within 440 px while either of them is in the cave, or anywhere during a hunt. It also comes after a player who hurts it from up to 1100 px away. It gives up after losing sight for 3 s, or after being drawn 1100 px from where the fight started. The cycle timer pauses while it fights the player.
- Foraging: it grazes cave food (`cave-feeding-*` patches) with its closest head, or wanders between the cave patches. Cave food gives it `HydraFoodXP` (0.6) XP per point of food *(tune)*. Kill loot gives only HP and MP.
- Hunting: it picks the nearest snake within 2600 px that it can safely beat (its HP × damage is at least 4× the snake's), or roams toward forest feeding spots. After one kill it eats the loot and heads home. Its spit and lunges hurt snakes, and a snake it attacks bites back for 6 s.
- The hydra and its cycle freeze while the player is not on floor 2, so it only starts growing once the player arrives.
- The boss bar at the top of the screen shows the level, the state (asleep, foraging, hunting, returning or enraged), the shared HP, and each head's HP, regrowth countdown or mana need. It appears while the hydra is fighting or is visible within 900 px. The minimap shows the hydra as a large purple dot.
- Feedback: a roar and screen shake when it wakes, a roar when a head regrows, a wet chop with a blood burst and "HEAD SEVERED" when a head is cut off, and a large burst with "THE HYDRA FALLS" when it dies. A sleeping hydra shows floating "z"s.
- Clues: three-toed tracks lead up all three approach paths and point the way into the cave tunnels. A shed skin lies in a clearing on each route, and bones are scattered around the lair.
- The location bar names the cave "II / THE SERPENT DEN".
- Testing playground: F10 resets the hydra and places the player inside the cave, and F11 gives the hydra 3 levels. `forestshot` with `SHOT_HYDRA=<ticks>` runs the real simulation in the cave and takes a screenshot. In this mode the two cell arguments set how close the player walks (shooting when above 0), and the target: 1 aims at the middle head and -1 does not shoot.

## Milestones
Build in this order. Each milestone should be playable on its own.

1. **Progression** *(done)*: the descent portal leads into floor 2 without ending the run; the win screen moves to a floor 3 placeholder.
2. **Forest map** *(in playtest)*: the new layout method, the forest itself, and navigation (minimap, fog of war, AI pathing).
3. **Basic snake** *(in playtest)*: segmented body, patrol routes, bite attack, respawn.
4. **Status effects** *(in playtest)*: the shared system, poison, and its HUD and visuals.
5. **Feeding and growth** *(in playtest)*: snakes eat, grow with rising costs, 20-segment cap.
6. **More snake attacks** *(in playtest)*: spit and fast slither.
7. **Forest perks** *(in playtest)*: the forest pool, ranks and the mixed offers at floor 2 statues.
8. **Cave and hydra** *(in playtest)*: cave clues, the cave arena, the hydra fight, head regrowth, feeding and hunting.
9. **Floor 3:** a small volcanic room reached through the hydra's teleport circle, ending the run with the win screen.

## Out of scope for v2
- Forest-specific food.
- Floor 3 content beyond a single small room.
- Progression that carries over between runs (meta-progression).

## Open questions
- Exact numbers, to find by playtesting: poison damage and tick rate, snake growth costs, perk values for each rank, the hydra's levels and cycle timings.
