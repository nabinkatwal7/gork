package main

type Item struct {
	ID         string
	Name       string
	Desc       string
	Type       string
	Slots      int
	Value      int
	Contraband bool
	Fruit      bool
}

type NPC struct {
	ID          string
	Name        string
	Desc        string
	Talk        string
	Disposition string
	Shop        []string
}

type Enemy struct {
	ID         string
	Name       string
	Desc       string
	HP         int
	MinDamage  int
	MaxDamage  int
	WantedGain int
	FleeChance float64
	IsBoss     bool
}

type Room struct {
	ID         string
	Name       string
	Island     string
	Desc       string
	Exits      map[string]string
	Items      []string
	NPCs       []string
	Enemies    []string
	Tags       []string
	CoordX     int
	CoordY     int
	Discovered bool
}

type Quest struct {
	ID      string
	Name    string
	Desc    string
	Active  bool
	Done    bool
	Outcome string
}

type Island struct {
	ID    string
	Name  string
	Desc  string
	Nodes []string
}

func WorldData() (map[string]*Room, map[string]*Item, map[string]*NPC, map[string]*Enemy, map[string]*Quest, map[string]*Island) {
	items := map[string]*Item{
		"rope":          {ID: "rope", Name: "Coil of Rope", Desc: "A trusty coil for daring entrances.", Type: "tool", Slots: 1, Value: 15},
		"flare":         {ID: "flare", Name: "Signal Flare", Desc: "A flare for emergencies or dramatic exits.", Type: "tool", Slots: 1, Value: 12},
		"nav_log":       {ID: "nav_log", Name: "Navigation Log", Desc: "A logbook full of winds, tides, and doodles.", Type: "quest", Slots: 1, Value: 35},
		"compass":       {ID: "compass", Name: "Brass Compass", Desc: "Points north and occasionally to snacks.", Type: "tool", Slots: 1, Value: 20},
		"grappling":     {ID: "grappling", Name: "Grappling Hook", Desc: "Hooky. Grippy. Dramatic.", Type: "tool", Slots: 1, Value: 25},
		"rum":           {ID: "rum", Name: "Bottle of Rum", Desc: "Liquid courage, corked tight.", Type: "consumable", Slots: 1, Value: 10},
		"spice":         {ID: "spice", Name: "Island Spice", Desc: "A pouch of spice with a fizzing aroma.", Type: "trade", Slots: 1, Value: 22},
		"bribe":         {ID: "bribe", Name: "Bribe Pouch", Desc: "Coins that clink with opportunity.", Type: "trade", Slots: 1, Value: 30, Contraband: true},
		"gadget_gull":   {ID: "gadget_gull", Name: "Wind-up Gull", Desc: "A mechanical gull that chirps on command.", Type: "tool", Slots: 1, Value: 28},
		"bounty_poster": {ID: "bounty_poster", Name: "Bounty Poster", Desc: "Someone else is wanted. That's reassuring.", Type: "lore", Slots: 1, Value: 5},
		"navy_badge":    {ID: "navy_badge", Name: "Bluecoat Badge", Desc: "A badge that screams 'confiscated'.", Type: "contraband", Slots: 1, Value: 40, Contraband: true},
		"cutlass":       {ID: "cutlass", Name: "Rusty Cutlass", Desc: "Seen more onions than battles.", Type: "weapon", Slots: 2, Value: 35},
		"flintlock":     {ID: "flintlock", Name: "Flintlock", Desc: "Old, loud, and still dangerous.", Type: "weapon", Slots: 2, Value: 60, Contraband: true},
		"chart":         {ID: "chart", Name: "Wild Current Chart", Desc: "A chart of the Wild Current routes.", Type: "quest", Slots: 1, Value: 40},
		"medkit":        {ID: "medkit", Name: "Med Kit", Desc: "Bandages, salve, and a lollipop.", Type: "consumable", Slots: 1, Value: 18},
		"balm":          {ID: "balm", Name: "Herbal Balm", Desc: "Smells like a forest after rain.", Type: "consumable", Slots: 1, Value: 15},
		"sun_coin":      {ID: "sun_coin", Name: "Sun Coin", Desc: "An ancient coin etched with a rising tide.", Type: "quest", Slots: 1, Value: 50},
		"stone_key":     {ID: "stone_key", Name: "Stone Key", Desc: "Heavy, carved with sea runes.", Type: "quest", Slots: 2, Value: 45},
		"cipher_lens":   {ID: "cipher_lens", Name: "Cipher Lens", Desc: "Reveals hidden script on Glyph Stones.", Type: "tool", Slots: 1, Value: 55},
		"glyph_frag_1":  {ID: "glyph_frag_1", Name: "Glyph Fragment A", Desc: "A fragment humming with old power.", Type: "quest", Slots: 1, Value: 0},
		"glyph_frag_2":  {ID: "glyph_frag_2", Name: "Glyph Fragment B", Desc: "A shard of carved stone.", Type: "quest", Slots: 1, Value: 0},
		"glyph_frag_3":  {ID: "glyph_frag_3", Name: "Glyph Fragment C", Desc: "The last fragment, warm to the touch.", Type: "quest", Slots: 1, Value: 0},
		"treasure_core": {ID: "treasure_core", Name: "Treasure Coordinate Core", Desc: "The legendary coordinates glow within.", Type: "quest", Slots: 1, Value: 0},
		"map_scrap":     {ID: "map_scrap", Name: "Map Scrap", Desc: "A torn scrap pointing inland.", Type: "lore", Slots: 1, Value: 8},
		"storm_lantern": {ID: "storm_lantern", Name: "Storm Lantern", Desc: "Refuses to go out, even in heavy rain.", Type: "tool", Slots: 1, Value: 20},
		"smoke_bomb":    {ID: "smoke_bomb", Name: "Smoke Bomb", Desc: "Great for exits. Also for excuses.", Type: "tool", Slots: 1, Value: 25},
		"sea_boots":     {ID: "sea_boots", Name: "Sea Boots", Desc: "Boots with weighted soles and great grip.", Type: "tool", Slots: 1, Value: 18},
		"pearl":         {ID: "pearl", Name: "Moon Pearl", Desc: "A luminous pearl with a cold glow.", Type: "trade", Slots: 1, Value: 45},
		"dock_pass":     {ID: "dock_pass", Name: "Dock Pass", Desc: "Lets you slip past port checks.", Type: "quest", Slots: 1, Value: 0},
		"repair_kit":    {ID: "repair_kit", Name: "Repair Kit", Desc: "Patchwork supplies for ship or gear.", Type: "tool", Slots: 1, Value: 20},
		"gale_fruit":    {ID: "gale_fruit", Name: "Gale Gale Fruit", Desc: "Swirls like a storm cloud.", Type: "fruit", Slots: 1, Value: 0, Fruit: true},
		"stone_fruit":   {ID: "stone_fruit", Name: "Stonewave Fruit", Desc: "Rumbles softly, like distant thunder.", Type: "fruit", Slots: 1, Value: 0, Fruit: true},
		"spark_fruit":   {ID: "spark_fruit", Name: "Sparkstep Fruit", Desc: "A crackling fruit that smells of rain.", Type: "fruit", Slots: 1, Value: 0, Fruit: true},
	}

	npcs := map[string]*NPC{
		"cook":       {ID: "cook", Name: "Ship Cook", Desc: "A cook with a ladle like a sword.", Talk: "Keep your hands busy and your belly fuller.", Disposition: "friendly"},
		"dockhand":   {ID: "dockhand", Name: "Dockhand", Desc: "A dockhand with a bandaged arm.", Talk: "Got any supplies? This arm's itching.", Disposition: "neutral"},
		"officer":    {ID: "officer", Name: "Bluecoat Officer", Desc: "A stern officer guarding the gate.", Talk: "Outpost access is restricted.", Disposition: "hostile"},
		"bartender":  {ID: "bartender", Name: "Tavern Bartender", Desc: "Polishing a mug with style.", Talk: "Rum loosens tongues and contracts.", Disposition: "neutral", Shop: []string{"rum", "smoke_bomb"}},
		"broker":     {ID: "broker", Name: "Shady Broker", Desc: "A broker with a grin that costs extra.", Talk: "Secrets are cheaper than anchors.", Disposition: "neutral", Shop: []string{"stone_key", "cipher_lens"}},
		"gadgeteer":  {ID: "gadgeteer", Name: "Gadgeteer", Desc: "Covered in soot and glitter.", Talk: "Spice makes my lenses sing.", Disposition: "neutral", Shop: []string{"gadget_gull", "storm_lantern"}},
		"herbalist":  {ID: "herbalist", Name: "Herbalist", Desc: "Sorting leaves with a smile.", Talk: "The jungle speaks if you listen.", Disposition: "friendly", Shop: []string{"balm", "medkit"}},
		"librarian":  {ID: "librarian", Name: "Mist Librarian", Desc: "A librarian with fog in her hair.", Talk: "Knowledge is safer when shared.", Disposition: "friendly"},
		"shipwright": {ID: "shipwright", Name: "Shipwright", Desc: "Wearing a belt of tools and sea salt.", Talk: "Fix the hull, fix the fate.", Disposition: "neutral", Shop: []string{"repair_kit", "sea_boots"}},
		"rival":      {ID: "rival", Name: "Rival Pirate", Desc: "A flashy pirate with a louder hat.", Talk: "The Wild Current has room for one legend.", Disposition: "hostile"},
		"priest":     {ID: "priest", Name: "Shrine Keeper", Desc: "Keeper of the storm shrine.", Talk: "Offerings calm the sky.", Disposition: "neutral"},
	}

	enemies := map[string]*Enemy{
		"reef_beast":   {ID: "reef_beast", Name: "Reef Beast", Desc: "A coral-covered brute with too many teeth.", HP: 14, MinDamage: 2, MaxDamage: 5, WantedGain: 0, FleeChance: 0.1},
		"navy_patrol":  {ID: "navy_patrol", Name: "Bluecoat Patrol", Desc: "Two Bluecoats with nets and attitude.", HP: 12, MinDamage: 2, MaxDamage: 4, WantedGain: 2, FleeChance: 0.2},
		"smuggler":     {ID: "smuggler", Name: "Spice Smuggler", Desc: "A smuggler guarding hidden crates.", HP: 10, MinDamage: 1, MaxDamage: 4, WantedGain: 1, FleeChance: 0.3},
		"rival_pirate": {ID: "rival_pirate", Name: "Rival Pirate", Desc: "A rival captain with a sharp grin.", HP: 16, MinDamage: 3, MaxDamage: 6, WantedGain: 2, FleeChance: 0.05, IsBoss: true},
		"navy_captain": {ID: "navy_captain", Name: "Bluecoat Captain", Desc: "A Navy captain with a polished saber.", HP: 18, MinDamage: 3, MaxDamage: 6, WantedGain: 3, FleeChance: 0.1, IsBoss: true},
	}

	rooms := map[string]*Room{
		"ship_deck":     {ID: "ship_deck", Name: "Rookie Deck", Island: "Ship", Desc: "Your scrappy ship bobs in the harbor. A note says: 'Try LOOK, INVENTORY, then GO NORTH.'", Exits: map[string]string{"north": "dock", "south": "ship_cabin"}, Items: []string{"rope", "flare"}, NPCs: []string{"cook"}, Tags: []string{"dock"}, CoordX: 2, CoordY: 2},
		"ship_cabin":    {ID: "ship_cabin", Name: "Captain's Cabin", Island: "Ship", Desc: "A cramped cabin with maps and ambition.", Exits: map[string]string{"north": "ship_deck"}, Items: []string{"nav_log", "compass"}, Tags: []string{}, CoordX: 2, CoordY: 3},
		"dock":          {ID: "dock", Name: "Harbor Dock", Island: "Harbor Isle", Desc: "Workers shout over gulls. The island town sprawls north.", Exits: map[string]string{"south": "ship_deck", "north": "town_square", "east": "market_lane", "west": "reef_shallows"}, Items: []string{"grappling"}, NPCs: []string{"dockhand"}, Tags: []string{"dock"}, CoordX: 2, CoordY: 1},
		"town_square":   {ID: "town_square", Name: "Town Square", Island: "Harbor Isle", Desc: "A plaza of stalls and gossip. A Bluecoat watches the gate.", Exits: map[string]string{"south": "dock", "east": "tavern", "west": "market_lane", "north": "navy_gate", "northeast": "shipyard"}, Items: []string{"bounty_poster"}, NPCs: []string{"officer"}, Tags: []string{}, CoordX: 2, CoordY: 0},
		"tavern":        {ID: "tavern", Name: "Tidal Tavern", Island: "Harbor Isle", Desc: "Sticky tables and loud rumors.", Exits: map[string]string{"west": "town_square"}, Items: []string{"rum"}, NPCs: []string{"bartender", "broker"}, Tags: []string{"shop"}, CoordX: 3, CoordY: 0},
		"market_lane":   {ID: "market_lane", Name: "Market Lane", Island: "Harbor Isle", Desc: "Lanterns sway over traders hawking gizmos.", Exits: map[string]string{"east": "town_square", "south": "dock", "west": "reef_shallows", "north": "jungle_path"}, Items: []string{"spice", "bribe", "gadget_gull"}, NPCs: []string{"gadgeteer"}, Tags: []string{"shop"}, CoordX: 1, CoordY: 0},
		"navy_gate":     {ID: "navy_gate", Name: "Bluecoat Gate", Island: "Harbor Isle", Desc: "A guarded gate leading to the Navy outpost.", Exits: map[string]string{"south": "town_square", "north": "navy_outpost"}, Items: []string{}, NPCs: []string{"officer"}, Tags: []string{}, CoordX: 2, CoordY: -1},
		"navy_outpost":  {ID: "navy_outpost", Name: "Bluecoat Outpost", Island: "Navy Bastion", Desc: "A stiff post of polished boots and judgment.", Exits: map[string]string{"south": "navy_gate"}, Items: []string{"navy_badge", "flintlock"}, Enemies: []string{"navy_captain"}, Tags: []string{"danger"}, CoordX: 2, CoordY: -2},
		"shipyard":      {ID: "shipyard", Name: "Shipyard", Island: "Harbor Isle", Desc: "Hull frames and resin scents fill the air.", Exits: map[string]string{"southwest": "town_square"}, Items: []string{"repair_kit", "sea_boots"}, NPCs: []string{"shipwright"}, Tags: []string{"shop"}, CoordX: 3, CoordY: -1},
		"reef_shallows": {ID: "reef_shallows", Name: "Reef Shallows", Island: "Harbor Isle", Desc: "Reefs glitter under the waves. The water looks deceptively calm.", Exits: map[string]string{"east": "dock", "north": "mist_pier"}, Items: []string{"gale_fruit"}, Enemies: []string{"reef_beast"}, Tags: []string{"danger"}, CoordX: 0, CoordY: 1},
		"jungle_path":   {ID: "jungle_path", Name: "Jungle Path", Island: "Ember Isle", Desc: "Vines twist like ropes. The ruins lie somewhere north.", Exits: map[string]string{"south": "market_lane", "north": "jungle_grove", "east": "ember_beach"}, Items: []string{"map_scrap"}, Tags: []string{}, CoordX: 1, CoordY: -1},
		"jungle_grove":  {ID: "jungle_grove", Name: "Jungle Grove", Island: "Ember Isle", Desc: "A grove with glowing fungus and a gentle breeze.", Exits: map[string]string{"south": "jungle_path", "north": "ruins_gate", "east": "ember_village"}, Items: []string{"medkit", "balm"}, NPCs: []string{"herbalist"}, Tags: []string{}, CoordX: 1, CoordY: -2},
		"ember_beach":   {ID: "ember_beach", Name: "Ember Beach", Island: "Ember Isle", Desc: "Black sand sparkles with heat.", Exits: map[string]string{"west": "jungle_path", "north": "ember_forge"}, Items: []string{"stone_fruit"}, Tags: []string{"danger"}, CoordX: 2, CoordY: -1},
		"ember_village": {ID: "ember_village", Name: "Ember Village", Island: "Ember Isle", Desc: "A village of smokehouses and laughter.", Exits: map[string]string{"west": "jungle_grove", "east": "ember_forge"}, Items: []string{"pearl"}, NPCs: []string{"priest"}, Tags: []string{"shop"}, CoordX: 2, CoordY: -2},
		"ember_forge":   {ID: "ember_forge", Name: "Ember Forge", Island: "Ember Isle", Desc: "A forge that never cools, guarded by a smuggler.", Exits: map[string]string{"south": "ember_beach", "west": "ember_village", "north": "ruins_gate"}, Items: []string{"glyph_frag_1", "cutlass"}, Enemies: []string{"smuggler"}, Tags: []string{"quest", "danger"}, CoordX: 2, CoordY: -3},
		"ruins_gate":    {ID: "ruins_gate", Name: "Ruins Gate", Island: "Ember Isle", Desc: "A stone gate carved with a riddle: 'Speak the sea and the stone will hear.'", Exits: map[string]string{"south": "jungle_grove", "north": "ruins_hall"}, Items: []string{"sun_coin"}, Tags: []string{"quest"}, CoordX: 1, CoordY: -3},
		"ruins_hall":    {ID: "ruins_hall", Name: "Glyph Hall", Island: "Ember Isle", Desc: "Dusty pillars and faded carvings.", Exits: map[string]string{"south": "ruins_gate", "north": "ruins_core"}, Items: []string{"glyph_frag_2"}, Tags: []string{"quest"}, CoordX: 1, CoordY: -4},
		"ruins_core":    {ID: "ruins_core", Name: "Glyph Core", Island: "Ember Isle", Desc: "A sealed chamber humming with the ocean's memory.", Exits: map[string]string{"south": "ruins_hall"}, Items: []string{}, Enemies: []string{"rival_pirate"}, Tags: []string{"quest", "danger"}, CoordX: 1, CoordY: -5},
		"mist_pier":     {ID: "mist_pier", Name: "Mist Pier", Island: "Mist Isle", Desc: "Fog rolls off the pier like breath.", Exits: map[string]string{"south": "reef_shallows", "north": "mist_library", "east": "mist_market"}, Items: []string{"spark_fruit"}, Tags: []string{"dock"}, CoordX: -1, CoordY: 1},
		"mist_library":  {ID: "mist_library", Name: "Mist Library", Island: "Mist Isle", Desc: "Shelves of scrolls whisper in the fog.", Exits: map[string]string{"south": "mist_pier"}, Items: []string{"glyph_frag_3"}, NPCs: []string{"librarian"}, Tags: []string{"quest"}, CoordX: -1, CoordY: 0},
		"mist_market":   {ID: "mist_market", Name: "Mist Market", Island: "Mist Isle", Desc: "Stalls glow with bioluminescent wares.", Exits: map[string]string{"west": "mist_pier"}, Items: []string{"smoke_bomb"}, Tags: []string{"shop"}, CoordX: 0, CoordY: 1},
		"sky_lift":      {ID: "sky_lift", Name: "Sky Lift", Island: "Skyline Atoll", Desc: "A lift platform rising toward the clouds.", Exits: map[string]string{"south": "mist_pier", "north": "sky_shrine"}, Items: []string{"chart"}, Tags: []string{"quest"}, CoordX: -2, CoordY: 0},
		"sky_shrine":    {ID: "sky_shrine", Name: "Sky Shrine", Island: "Skyline Atoll", Desc: "A shrine in the clouds, lightning crackling nearby.", Exits: map[string]string{"south": "sky_lift"}, Items: []string{"stone_key"}, NPCs: []string{"priest"}, Tags: []string{"quest"}, CoordX: -2, CoordY: -1},
	}

	quests := map[string]*Quest{
		"main":       {ID: "main", Name: "Glyph Stone Hunt", Desc: "Collect three Glyph Stone fragments and decipher their coordinates.", Active: true},
		"dockhand":   {ID: "dockhand", Name: "Bandaged Dockhand", Desc: "Help the dockhand and earn their trust.", Active: true},
		"gadgeteer":  {ID: "gadgeteer", Name: "Spice for Gadgets", Desc: "Trade spice for a cipher lens.", Active: true},
		"broker":     {ID: "broker", Name: "Rum for Keys", Desc: "Trade rum for a stone key.", Active: true},
		"priest":     {ID: "priest", Name: "Shrine Offering", Desc: "Bring a sun coin to the shrine keeper.", Active: true},
		"shipwright": {ID: "shipwright", Name: "Hull Repairs", Desc: "Deliver a repair kit for a dock pass.", Active: true},
		"rival":      {ID: "rival", Name: "Rival Showdown", Desc: "Defeat the rival pirate in the ruins.", Active: true},
	}

	islands := map[string]*Island{
		"Harbor Isle":   {ID: "Harbor Isle", Name: "Harbor Isle", Desc: "A bustling island of trade and gossip.", Nodes: []string{"dock", "town_square", "tavern", "market_lane", "shipyard"}},
		"Ember Isle":    {ID: "Ember Isle", Name: "Ember Isle", Desc: "A volcanic island with ancient ruins.", Nodes: []string{"jungle_path", "jungle_grove", "ember_beach", "ember_village", "ember_forge", "ruins_gate", "ruins_hall", "ruins_core"}},
		"Mist Isle":     {ID: "Mist Isle", Name: "Mist Isle", Desc: "An island cloaked in gentle fog.", Nodes: []string{"mist_pier", "mist_library", "mist_market"}},
		"Skyline Atoll": {ID: "Skyline Atoll", Name: "Skyline Atoll", Desc: "A cloud-touched atoll of storms.", Nodes: []string{"sky_lift", "sky_shrine"}},
		"Navy Bastion":  {ID: "Navy Bastion", Name: "Navy Bastion", Desc: "The Bluecoat Navy stronghold.", Nodes: []string{"navy_outpost"}},
		"Ship":          {ID: "Ship", Name: "Ship", Desc: "Your vessel and home.", Nodes: []string{"ship_deck", "ship_cabin"}},
	}

	return rooms, items, npcs, enemies, quests, islands
}
