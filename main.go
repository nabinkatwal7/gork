package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Game struct {
	Rooms      map[string]*Room
	Items      map[string]*Item
	NPCs       map[string]*NPC
	Enemies    map[string]*Enemy
	Player     Player
	Flags      map[string]bool
	Wanted     int
	Reputation int
	Running    bool
}

type Player struct {
	Location  string
	Inventory []string
	MaxSlots  int
	HP        int
	HasFruit  bool
}

type Room struct {
	Name    string
	Desc    string
	Exits   map[string]string
	Items   []string
	NPCs    []string
	Enemies []string
}

type Item struct {
	Name    string
	Desc    string
	Size    int
	Illegal bool
}

type NPC struct {
	Name string
	Desc string
	Talk string
}

type Enemy struct {
	Name string
	Desc string
	HP   int
}

type SaveData struct {
	Player      Player
	RoomItems   map[string][]string
	RoomEnemies map[string][]string
	EnemyHP     map[string]int
	Flags       map[string]bool
	Wanted      int
	Reputation  int
}

func main() {
	rand.Seed(time.Now().UnixNano())
	game := NewGame()
	game.Intro()
	game.Loop()
}

func NewGame() *Game {
	items := map[string]*Item{
		"rope":         {Name: "coil of rope", Desc: "A salty coil, perfect for boarding or nonsense.", Size: 1},
		"flare":        {Name: "signal flare", Desc: "A single flare that screams 'rescue me' or 'chase me'.", Size: 1},
		"nav_log":      {Name: "navigation log", Desc: "A soggy logbook full of wind notes and doodles.", Size: 1},
		"compass":      {Name: "brass compass", Desc: "It points north and occasionally to snacks.", Size: 1},
		"grappling":    {Name: "grappling hook", Desc: "Hooky. Grippy. Dramatic.", Size: 1},
		"conch":        {Name: "conch whistle", Desc: "Blow it and the sea seems to listen.", Size: 1},
		"rum":          {Name: "bottle of rum", Desc: "Liquid courage with a pirate label.", Size: 1},
		"spice":        {Name: "spice pouch", Desc: "A pouch of island spice, priceless to tinkerers.", Size: 1},
		"bribe":        {Name: "bribe pouch", Desc: "Coins that clink with opportunity.", Size: 1},
		"gull":         {Name: "wind-up gull", Desc: "A gadget gull that chirps when you crank it.", Size: 1},
		"bounty":       {Name: "bounty poster", Desc: "A poster of some unlucky pirate. Not you. Yet.", Size: 1},
		"navy_badge":   {Name: "navy badge", Desc: "A Bluecoat Navy badge. Possessing it is a bad idea.", Size: 1, Illegal: true},
		"cutlass":      {Name: "rusty cutlass", Desc: "Seen more onions than battles.", Size: 2},
		"flintlock":    {Name: "flintlock", Desc: "Old, loud, and still dangerous.", Size: 2, Illegal: true},
		"chart":        {Name: "sea chart", Desc: "A chart labeled 'Wild Current'.", Size: 1},
		"medkit":       {Name: "med kit", Desc: "Bandages, salve, and a lollipop.", Size: 1},
		"balm":         {Name: "herbal balm", Desc: "Smells like a forest after rain.", Size: 1},
		"cursed_fruit": {Name: "Gale-Gale Cursed Fruit", Desc: "Swirls like a storm cloud. It dares you to bite.", Size: 1},
		"sun_coin":     {Name: "sun coin", Desc: "An ancient coin etched with a rising tide.", Size: 1},
		"stone_key":    {Name: "stone key", Desc: "Heavy, carved with sea runes.", Size: 2},
		"cipher_lens":  {Name: "cipher lens", Desc: "A lens that reveals hidden script.", Size: 1},
		"coords":       {Name: "sky-etched coordinates", Desc: "The legendary treasure coordinates, glittering with promise.", Size: 1},
		"map_scrap":    {Name: "map scrap", Desc: "A torn scrap showing jungle paths.", Size: 1},
		"lantern":      {Name: "storm lantern", Desc: "A lantern that refuses to go out.", Size: 1},
	}

	rooms := map[string]*Room{
		"ship_deck": {
			Name:  "Rookie Deck",
			Desc:  "Your scrappy ship bobs in the harbor. A handwritten note says: 'Try LOOK, INVENTORY, then GO NORTH.'",
			Exits: map[string]string{"north": "dock", "south": "ship_cabin"},
			Items: []string{"rope", "flare"},
			NPCs:  []string{"cook"},
		},
		"ship_cabin": {
			Name:  "Captain's Cabin",
			Desc:  "A cramped cabin with maps, smells, and ambition.",
			Exits: map[string]string{"north": "ship_deck"},
			Items: []string{"nav_log", "compass"},
			NPCs:  []string{"old_sailor"},
		},
		"dock": {
			Name:  "Harbor Dock",
			Desc:  "Workers shout over gulls. The island town sprawls north.",
			Exits: map[string]string{"south": "ship_deck", "north": "town_square", "east": "market_lane", "west": "reef_shallows"},
			Items: []string{"grappling", "conch"},
			NPCs:  []string{"dockhand"},
		},
		"town_square": {
			Name:  "Town Square",
			Desc:  "A plaza of stalls and gossip. A Bluecoat officer watches the northern gate.",
			Exits: map[string]string{"south": "dock", "east": "tavern", "west": "market_lane", "north": "navy_outpost"},
			Items: []string{"bounty"},
			NPCs:  []string{"officer"},
		},
		"tavern": {
			Name:  "Tidal Tavern",
			Desc:  "A tavern with sticky tables and louder rumors.",
			Exits: map[string]string{"west": "town_square"},
			Items: []string{"rum"},
			NPCs:  []string{"bartender", "broker"},
		},
		"market_lane": {
			Name:  "Market Lane",
			Desc:  "Lanterns sway over traders hawking gizmos and spices.",
			Exits: map[string]string{"west": "town_square", "east": "jungle_path", "south": "dock", "north": "signal_hill"},
			Items: []string{"spice", "bribe", "gull"},
			NPCs:  []string{"gadgeteer"},
		},
		"navy_outpost": {
			Name:    "Bluecoat Outpost",
			Desc:    "A stiff post of polished boots and judgment.",
			Exits:   map[string]string{"south": "town_square"},
			Items:   []string{"navy_badge", "cutlass", "flintlock"},
			Enemies: []string{"navy_captain"},
		},
		"jungle_path": {
			Name:  "Jungle Path",
			Desc:  "Vines twist like ropes. The ruins lie somewhere north.",
			Exits: map[string]string{"west": "market_lane", "east": "jungle_clearing", "north": "ruins_gate"},
			Items: []string{"map_scrap"},
		},
		"jungle_clearing": {
			Name:  "Jungle Clearing",
			Desc:  "A clearing with medicinal plants and humming insects.",
			Exits: map[string]string{"west": "jungle_path", "north": "signal_hill"},
			Items: []string{"medkit", "balm"},
			NPCs:  []string{"herbalist"},
		},
		"signal_hill": {
			Name:  "Signal Hill",
			Desc:  "A windy hill with a view of the Wild Current.",
			Exits: map[string]string{"south": "market_lane", "east": "jungle_clearing"},
			Items: []string{"chart", "lantern"},
		},
		"reef_shallows": {
			Name:    "Reef Shallows",
			Desc:    "Reefs glitter under the waves. The water looks deceptively calm.",
			Exits:   map[string]string{"east": "dock"},
			Items:   []string{"cursed_fruit"},
			Enemies: []string{"reef_beast"},
		},
		"ruins_gate": {
			Name:  "Ruins Gate",
			Desc:  "A stone gate carved with a riddle: 'Speak the sea and the stone will hear.'",
			Exits: map[string]string{"south": "jungle_path", "north": "ancient_ruin"},
			Items: []string{"sun_coin"},
			NPCs:  []string{"hermit"},
		},
		"ancient_ruin": {
			Name:  "Ancient Ruin",
			Desc:  "Dusty pillars hold a ceiling painted with storms.",
			Exits: map[string]string{"south": "ruins_gate", "north": "inner_chamber"},
		},
		"inner_chamber": {
			Name:  "Inner Chamber",
			Desc:  "Glyphs glow faintly. A pedestal waits for a daring hand.",
			Exits: map[string]string{"south": "ancient_ruin"},
			Items: []string{"coords"},
		},
	}

	npcs := map[string]*NPC{
		"cook":       {Name: "ship cook", Desc: "A cook sharpening a spoon.", Talk: "Keep your hands busy and your belly fuller."},
		"old_sailor": {Name: "old sailor", Desc: "An old sailor with a sea-glass eye.", Talk: "Storms respect those who respect storms."},
		"dockhand":   {Name: "dockhand", Desc: "A dockhand with a bandaged arm.", Talk: "Got any supplies? This arm's itching."},
		"officer":    {Name: "bluecoat officer", Desc: "A stern officer guarding the north gate.", Talk: "Outpost access is restricted. Papers or payment."},
		"bartender":  {Name: "bartender", Desc: "A bartender polishing a mug.", Talk: "Rum loosens tongues and contracts."},
		"broker":     {Name: "shady broker", Desc: "A broker with a grin that costs extra.", Talk: "I trade in secrets and keys. Bring me a bottle."},
		"gadgeteer":  {Name: "gadgeteer", Desc: "A gadgeteer covered in soot.", Talk: "Spice makes my lenses sing."},
		"herbalist":  {Name: "herbalist", Desc: "A gentle herbalist sorting leaves.", Talk: "The jungle speaks if you listen."},
		"hermit":     {Name: "riddle hermit", Desc: "A hermit humming sea shanties.", Talk: "The gate likes the sea's voice. Loud and clear."},
	}

	enemies := map[string]*Enemy{
		"reef_beast":   {Name: "reef beast", Desc: "A coral-covered brute with too many teeth.", HP: 12},
		"navy_captain": {Name: "navy captain", Desc: "A Bluecoat captain with a polished saber.", HP: 14},
		"navy_patrol":  {Name: "navy patrol", Desc: "A pair of Bluecoats with nets and attitude.", HP: 10},
	}

	return &Game{
		Rooms:      rooms,
		Items:      items,
		NPCs:       npcs,
		Enemies:    enemies,
		Player:     Player{Location: "ship_deck", Inventory: []string{}, MaxSlots: 8, HP: 20},
		Flags:      map[string]bool{},
		Wanted:     0,
		Reputation: 0,
		Running:    true,
	}
}

func (g *Game) Intro() {
	fmt.Println("=== Wild Current: A Pirate Tale ===")
	fmt.Println("You are a rookie captain chasing legendary treasure across the Wild Current.")
	fmt.Println("Type HELP for commands. Try LOOK, INVENTORY, and GO NORTH to begin.")
	fmt.Println()
	g.Look()
}

func (g *Game) Loop() {
	reader := bufio.NewReader(os.Stdin)
	for g.Running {
		fmt.Print("\n> ")
		line, _ := reader.ReadString('\n')
		g.HandleCommand(strings.TrimSpace(line))
	}
}

func (g *Game) HandleCommand(input string) {
	if input == "" {
		return
	}
	parts := strings.Fields(strings.ToLower(input))
	if len(parts) == 0 {
		return
	}

	verb := parts[0]
	if dir := normalizeDir(verb); dir != "" {
		g.Move(dir)
		return
	}

	switch verb {
	case "go", "move":
		if len(parts) < 2 {
			fmt.Println("Go where?")
			return
		}
		g.Move(normalizeDir(parts[1]))
	case "look", "l":
		g.Look()
	case "examine", "x":
		if len(parts) < 2 {
			fmt.Println("Examine what?")
			return
		}
		g.Examine(strings.Join(parts[1:], " "))
	case "take", "get":
		if len(parts) < 2 {
			fmt.Println("Take what?")
			return
		}
		g.Take(strings.Join(parts[1:], " "))
	case "drop":
		if len(parts) < 2 {
			fmt.Println("Drop what?")
			return
		}
		g.Drop(strings.Join(parts[1:], " "))
	case "inventory", "i":
		g.Inventory()
	case "talk":
		if len(parts) < 2 {
			fmt.Println("Talk to whom?")
			return
		}
		g.Talk(strings.Join(parts[1:], " "))
	case "use":
		if len(parts) < 2 {
			fmt.Println("Use what?")
			return
		}
		g.Use(parts[1:], input)
	case "attack":
		if len(parts) < 2 {
			fmt.Println("Attack whom?")
			return
		}
		g.Attack(strings.Join(parts[1:], " "))
	case "help":
		g.Help()
	case "save":
		g.Save("save.json")
	case "load":
		g.Load("save.json")
	case "quit", "exit":
		fmt.Println("You lower the sails and end your tale... for now.")
		g.Running = false
	default:
		fmt.Println("Unknown command. Type HELP for options.")
	}
}

func normalizeDir(dir string) string {
	switch dir {
	case "north", "n":
		return "north"
	case "south", "s":
		return "south"
	case "east", "e":
		return "east"
	case "west", "w":
		return "west"
	default:
		return ""
	}
}

func (g *Game) Look() {
	room := g.Rooms[g.Player.Location]
	fmt.Printf("%s\n%s\n", room.Name, room.Desc)
	if len(room.Items) > 0 {
		fmt.Println("You see:")
		for _, itemID := range room.Items {
			fmt.Printf("- %s\n", g.Items[itemID].Name)
		}
	}
	if len(room.NPCs) > 0 {
		fmt.Println("People here:")
		for _, npcID := range room.NPCs {
			fmt.Printf("- %s\n", g.NPCs[npcID].Name)
		}
	}
	if len(room.Enemies) > 0 {
		fmt.Println("Threats present:")
		for _, enemyID := range room.Enemies {
			fmt.Printf("- %s\n", g.Enemies[enemyID].Name)
		}
	}
	fmt.Printf("Exits: %s\n", strings.Join(exitList(room.Exits), ", "))
}

func exitList(exits map[string]string) []string {
	keys := make([]string, 0, len(exits))
	for key := range exits {
		keys = append(keys, key)
	}
	return keys
}

func (g *Game) Move(direction string) {
	if direction == "" {
		fmt.Println("That direction makes no sense.")
		return
	}
	room := g.Rooms[g.Player.Location]
	dest, ok := room.Exits[direction]
	if !ok {
		fmt.Println("You can't go that way.")
		return
	}

	if !g.CanEnter(dest) {
		return
	}

	g.Player.Location = dest
	g.Look()
	g.MaybePatrol()
}

func (g *Game) CanEnter(dest string) bool {
	if dest == "navy_outpost" && !g.Flags["bribed"] {
		fmt.Println("The bluecoat officer blocks the way. A 'donation' might help.")
		return false
	}
	if dest == "ruins_gate" && !g.Flags["mapDecoded"] {
		fmt.Println("The jungle splits endlessly. You need better directions.")
		return false
	}
	if dest == "ancient_ruin" && !g.Flags["ruinUnlocked"] {
		fmt.Println("The stone gate is locked tight.")
		return false
	}
	if dest == "inner_chamber" && !g.Flags["innerUnlocked"] {
		fmt.Println("A sealed stone door bars the way. The riddle must be answered.")
		return false
	}
	if dest == "reef_shallows" && g.Player.HasFruit {
		g.EndGame("The cursed power drags you under the waves. The sea takes its price.")
		return false
	}
	if g.Wanted >= 5 && (dest == "navy_outpost" || dest == "town_square") {
		g.EndGame("Bluecoat Navy surrounds you. You are captured and hauled away.")
		return false
	}
	return true
}

func (g *Game) Examine(name string) {
	if itemID := g.FindItem(name, g.Player.Inventory); itemID != "" {
		fmt.Println(g.Items[itemID].Desc)
		return
	}
	room := g.Rooms[g.Player.Location]
	if itemID := g.FindItem(name, room.Items); itemID != "" {
		fmt.Println(g.Items[itemID].Desc)
		return
	}
	if npcID := g.FindNPC(name, room.NPCs); npcID != "" {
		fmt.Println(g.NPCs[npcID].Desc)
		return
	}
	if enemyID := g.FindEnemy(name, room.Enemies); enemyID != "" {
		fmt.Println(g.Enemies[enemyID].Desc)
		return
	}
	fmt.Println("You find nothing like that to examine.")
}

func (g *Game) Take(name string) {
	room := g.Rooms[g.Player.Location]
	itemID := g.FindItem(name, room.Items)
	if itemID == "" {
		fmt.Println("You don't see that here.")
		return
	}
	if !g.CanCarry(itemID) {
		fmt.Println("You're carrying too much already.")
		return
	}
	room.Items = removeID(room.Items, itemID)
	g.Player.Inventory = append(g.Player.Inventory, itemID)
	fmt.Printf("You take the %s.\n", g.Items[itemID].Name)
	if g.Items[itemID].Illegal {
		g.Wanted++
		fmt.Println("That felt illegal. Wanted level rises.")
	}
}

func (g *Game) Drop(name string) {
	itemID := g.FindItem(name, g.Player.Inventory)
	if itemID == "" {
		fmt.Println("You don't have that.")
		return
	}
	g.Player.Inventory = removeID(g.Player.Inventory, itemID)
	g.Rooms[g.Player.Location].Items = append(g.Rooms[g.Player.Location].Items, itemID)
	fmt.Printf("You drop the %s.\n", g.Items[itemID].Name)
}

func (g *Game) Inventory() {
	if len(g.Player.Inventory) == 0 {
		fmt.Println("Your pockets are empty.")
		return
	}
	fmt.Printf("You carry (%d/%d slots):\n", g.InventorySlots(), g.Player.MaxSlots)
	for _, itemID := range g.Player.Inventory {
		fmt.Printf("- %s\n", g.Items[itemID].Name)
	}
}

func (g *Game) Talk(name string) {
	room := g.Rooms[g.Player.Location]
	npcID := g.FindNPC(name, room.NPCs)
	if npcID == "" {
		fmt.Println("No one like that is here.")
		return
	}
	fmt.Println(g.NPCs[npcID].Talk)
}

func (g *Game) Use(args []string, raw string) {
	lower := strings.ToLower(raw)
	parts := strings.Split(lower, " on ")
	itemName := strings.TrimSpace(strings.TrimPrefix(parts[0], "use "))
	target := ""
	if len(parts) > 1 {
		target = strings.TrimSpace(parts[1])
	}
	itemID := g.FindItem(itemName, g.Player.Inventory)
	if itemID == "" {
		fmt.Println("You don't have that to use.")
		return
	}

	switch itemID {
	case "rum":
		if target == "shady broker" || target == "broker" {
			if g.Flags["brokerPaid"] {
				fmt.Println("The broker already got their drink.")
				return
			}
			g.Flags["brokerPaid"] = true
			g.Player.Inventory = removeID(g.Player.Inventory, "rum")
			g.Player.Inventory = append(g.Player.Inventory, "stone_key")
			fmt.Println("The broker trades the rum for a stone key.")
			return
		}
	case "spice":
		if target == "gadgeteer" {
			if g.Flags["gadgeteerTraded"] {
				fmt.Println("The gadgeteer already gave you their best lens.")
				return
			}
			g.Flags["gadgeteerTraded"] = true
			g.Player.Inventory = removeID(g.Player.Inventory, "spice")
			g.Player.Inventory = append(g.Player.Inventory, "cipher_lens")
			fmt.Println("The gadgeteer trades a cipher lens for the spice.")
			return
		}
	case "medkit":
		if target == "dockhand" {
			if g.Flags["dockhandHelped"] {
				fmt.Println("The dockhand already patched up.")
				return
			}
			g.Flags["dockhandHelped"] = true
			g.Reputation++
			g.Player.Inventory = removeID(g.Player.Inventory, "medkit")
			g.Player.Inventory = append(g.Player.Inventory, "sun_coin")
			fmt.Println("You patch the dockhand. They slip you an ancient sun coin.")
			return
		}
	case "bribe":
		if target == "bluecoat officer" || target == "officer" {
			if g.Flags["bribed"] {
				fmt.Println("The officer already turned a blind eye.")
				return
			}
			g.Flags["bribed"] = true
			g.Player.Inventory = removeID(g.Player.Inventory, "bribe")
			fmt.Println("The officer pockets the coins and steps aside.")
			return
		}
	case "cipher_lens":
		if target == "navigation log" || target == "nav log" || target == "log" {
			if !g.HasItem("nav_log") {
				fmt.Println("You need the navigation log to decipher.")
				return
			}
			g.Flags["mapDecoded"] = true
			fmt.Println("The cipher lens reveals a route to the jungle ruins.")
			return
		}
	case "stone_key":
		if target == "ruin gate" || target == "gate" {
			g.Flags["ruinUnlocked"] = true
			fmt.Println("The stone key turns. The gate groans open.")
			return
		}
	case "conch":
		if target == "stone door" || target == "door" {
			g.Flags["innerUnlocked"] = true
			fmt.Println("The conch's call echoes. The stone door slides aside.")
			return
		}
	case "cursed_fruit":
		g.Player.HasFruit = true
		g.Player.Inventory = removeID(g.Player.Inventory, "cursed_fruit")
		fmt.Println("Power surges through you! Wind answers your gestures, but the sea now hates you.")
		return
	case "coords":
		if target == "helm" || target == "ship" || target == "wheel" {
			g.EndingWithTreasure()
			return
		}
	}

	fmt.Println("Nothing happens.")
}

func (g *Game) Attack(name string) {
	room := g.Rooms[g.Player.Location]
	enemyID := g.FindEnemy(name, room.Enemies)
	if enemyID == "" {
		fmt.Println("No enemy by that name is here.")
		return
	}
	enemy := g.Enemies[enemyID]
	fmt.Printf("You lunge at the %s!\n", enemy.Name)
	playerHit := rand.Float64() < 0.65
	if g.Player.HasFruit {
		playerHit = rand.Float64() < 0.8
	}
	if playerHit {
		damage := rand.Intn(4) + 2
		enemy.HP -= damage
		fmt.Printf("You strike for %d damage.\n", damage)
	} else {
		fmt.Println("You miss and stumble.")
	}

	if enemy.HP <= 0 {
		fmt.Printf("The %s collapses.\n", enemy.Name)
		room.Enemies = removeID(room.Enemies, enemyID)
		if enemyID == "navy_captain" || enemyID == "navy_patrol" {
			g.Wanted += 2
			fmt.Println("The Navy will not forget this.")
		}
		return
	}

	enemyHit := rand.Float64() < 0.5
	if enemyHit {
		damage := rand.Intn(3) + 1
		g.Player.HP -= damage
		fmt.Printf("The %s hits you for %d damage. (HP %d)\n", enemy.Name, damage, g.Player.HP)
	} else {
		fmt.Printf("The %s swings wide.\n", enemy.Name)
	}

	if g.Player.HP <= 0 {
		g.EndGame("You slump to the ground. The Bluecoat Navy captures you.")
	}
}

func (g *Game) Help() {
	fmt.Println("Commands:")
	fmt.Println("- Movement: GO NORTH, NORTH, N (also SOUTH/EAST/WEST)")
	fmt.Println("- Actions: LOOK, EXAMINE <thing>, TAKE <item>, DROP <item>")
	fmt.Println("- Inventory: INVENTORY or I")
	fmt.Println("- Interactions: TALK <npc>, USE <item> [ON <target>], ATTACK <enemy>")
	fmt.Println("- Utility: HELP, SAVE, LOAD, QUIT")
	fmt.Println("Goal: Find the sky-etched coordinates in the ancient ruins and escape.")
	fmt.Println("Watch your Wanted Level and Reputation. Cursed Fruits grant power with a price.")
}

func (g *Game) Save(filename string) {
	data := SaveData{
		Player:      g.Player,
		RoomItems:   map[string][]string{},
		RoomEnemies: map[string][]string{},
		EnemyHP:     map[string]int{},
		Flags:       g.Flags,
		Wanted:      g.Wanted,
		Reputation:  g.Reputation,
	}
	for id, room := range g.Rooms {
		data.RoomItems[id] = append([]string{}, room.Items...)
		data.RoomEnemies[id] = append([]string{}, room.Enemies...)
	}
	for id, enemy := range g.Enemies {
		data.EnemyHP[id] = enemy.HP
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Could not save game:", err)
		return
	}
	if err := os.WriteFile(filename, raw, 0644); err != nil {
		fmt.Println("Could not write save file:", err)
		return
	}
	fmt.Println("Game saved to", filename)
}

func (g *Game) Load(filename string) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Could not load save file:", err)
		return
	}
	var data SaveData
	if err := json.Unmarshal(raw, &data); err != nil {
		fmt.Println("Save file corrupted:", err)
		return
	}
	fresh := NewGame()
	*g = *fresh
	g.Player = data.Player
	g.Flags = data.Flags
	g.Wanted = data.Wanted
	g.Reputation = data.Reputation
	for id, items := range data.RoomItems {
		if room, ok := g.Rooms[id]; ok {
			room.Items = items
		}
	}
	for id, enemies := range data.RoomEnemies {
		if room, ok := g.Rooms[id]; ok {
			room.Enemies = enemies
		}
	}
	for id, hp := range data.EnemyHP {
		if enemy, ok := g.Enemies[id]; ok {
			enemy.HP = hp
		}
	}
	fmt.Println("Game loaded.")
	g.Look()
}

func (g *Game) MaybePatrol() {
	if g.Wanted < 3 {
		return
	}
	room := g.Rooms[g.Player.Location]
	if room == nil || len(room.Enemies) > 0 {
		return
	}
	if g.Player.Location == "ship_deck" || g.Player.Location == "ship_cabin" {
		return
	}
	if rand.Float64() < 0.35 {
		room.Enemies = append(room.Enemies, "navy_patrol")
		fmt.Println("A Bluecoat Navy patrol storms in, nets ready.")
	}
}

func (g *Game) EndingWithTreasure() {
	if g.Player.Location != "ship_deck" {
		fmt.Println("You need to be on your ship's helm to use the coordinates.")
		return
	}
	if g.Player.HasFruit {
		g.EndGame("You escape with the treasure, but the cursed power twists your fate. The sea will always hunt you.")
		return
	}
	if g.Wanted >= 4 {
		g.EndGame("You reach the helm, but the Bluecoat Navy intercepts you. Captured with treasure in hand.")
		return
	}
	g.EndGame("You steer into the Wild Current and vanish with the treasure. Legends will whisper your name.")
}

func (g *Game) EndGame(message string) {
	fmt.Println("\n=== The End ===")
	fmt.Println(message)
	g.Running = false
}

func (g *Game) FindItem(name string, list []string) string {
	name = strings.ToLower(name)
	for _, itemID := range list {
		item := g.Items[itemID]
		if strings.ToLower(item.Name) == name || itemID == name {
			return itemID
		}
	}
	return ""
}

func (g *Game) FindNPC(name string, list []string) string {
	name = strings.ToLower(name)
	for _, npcID := range list {
		npc := g.NPCs[npcID]
		if strings.ToLower(npc.Name) == name || npcID == name {
			return npcID
		}
	}
	return ""
}

func (g *Game) FindEnemy(name string, list []string) string {
	name = strings.ToLower(name)
	for _, enemyID := range list {
		enemy := g.Enemies[enemyID]
		if strings.ToLower(enemy.Name) == name || enemyID == name {
			return enemyID
		}
	}
	return ""
}

func removeID(list []string, id string) []string {
	result := make([]string, 0, len(list))
	for _, entry := range list {
		if entry != id {
			result = append(result, entry)
		}
	}
	return result
}

func (g *Game) InventorySlots() int {
	total := 0
	for _, itemID := range g.Player.Inventory {
		total += g.Items[itemID].Size
	}
	return total
}

func (g *Game) CanCarry(itemID string) bool {
	return g.InventorySlots()+g.Items[itemID].Size <= g.Player.MaxSlots
}

func (g *Game) HasItem(itemID string) bool {
	for _, id := range g.Player.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}
